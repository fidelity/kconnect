/*
Copyright 2020 The kconnect Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/versent/saml2aws"
	"github.com/versent/saml2aws/pkg/awsconfig"
	"github.com/versent/saml2aws/pkg/cfg"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/identity"
)

const (
	responseTag = "Response"
)

var (
	ErrNoRegion               = errors.New("no region found")
	ErrNoRolesFound           = errors.New("no aws roles found")
	ErrNotAccounts            = errors.New("no accounts available")
	ErrMissingResponseElement = errors.New("missing response element")
)

type awsProviderConfig struct {
	sp.ProviderConfig

	Partition string `json:"partition" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

func NewServiceProvider(itemSelector provider.SelectItemFunc) sp.ServiceProvider {
	return &ServiceProvider{
		logger:       zap.S().With("provider", "saml", "sp", "aws"),
		itemSelector: itemSelector,
	}
}

type ServiceProvider struct {
	logger       *zap.SugaredLogger
	itemSelector provider.SelectItemFunc
}

func (p *ServiceProvider) PopulateAccount(account *cfg.IDPAccount, cfg config.ConfigurationSet) error {
	account.AmazonWebservicesURN = "urn:amazon:webservices"
	account.Profile = "kconnect-saml-provider"

	regionCfg := cfg.Get("region")
	if regionCfg == nil || regionCfg.Value.(string) == "" {
		return ErrNoRegion
	}
	account.Region = regionCfg.Value.(string)

	roleCfg := cfg.Get("role-arn")
	if roleCfg != nil || roleCfg.Value.(string) != "" {
		account.RoleARN = roleCfg.Value.(string)
	}

	return nil
}

func (p *ServiceProvider) ProcessAssertions(account *cfg.IDPAccount, samlAssertions string, cfg config.ConfigurationSet) (identity.Identity, error) {
	data, err := base64.StdEncoding.DecodeString(samlAssertions)
	if err != nil {
		return nil, fmt.Errorf("decoding SAMLAssertion: %w", err)
	}

	roles, err := saml2aws.ExtractAwsRoles(data)
	if err != nil {
		return nil, fmt.Errorf("extracting AWS roles from assertion: %w", err)
	}

	if len(roles) == 0 {
		return nil, ErrNoRolesFound
	}

	awsRoles, err := saml2aws.ParseAWSRoles(roles)
	if err != nil {
		return nil, fmt.Errorf("parsing aws roles: %w", err)
	}

	roleFilter := ""
	if cfg.ExistsWithValue("role-filter") {
		item := cfg.Get("role-filter")
		roleFilter = item.Value.(string)
	}

	role, err := p.resolveRole(awsRoles, samlAssertions, account, roleFilter)
	if err != nil {
		return nil, fmt.Errorf("resolving aws role: %w", err)
	}

	if err := cfg.SetValue("role-arn", role.RoleARN); err != nil {
		return nil, fmt.Errorf("setting role-arn config value: %w", err)
	}
	p.logger.Debugw("role selected", "role", role.RoleARN)

	awsCreds, err := p.loginToStsUsingRole(account, role, samlAssertions)
	if err != nil {
		return nil, fmt.Errorf("logging into AWS using STS and SAMLAssertion: %w", err)
	}

	// Create profile based on the AWS creds
	identifier, err := kaws.CreateIDFromCreds(awsCreds)
	if err != nil {
		return nil, fmt.Errorf("creating identifier from AWS creds: %w", err)
	}
	profileName := fmt.Sprintf("kconnect-%s", identifier)
	if err := p.setProfileName(profileName, cfg); err != nil {
		return nil, fmt.Errorf("setting profile name: %w", err)
	}

	awsSharedCredentialsFile := ""
	if cfg.ExistsWithValue("aws-shared-credentials-file") {
		item := cfg.Get("aws-shared-credentials-file")
		awsSharedCredentialsFile = item.Value.(string)
	}

	awsIdentity := kaws.MapCredsToIdentity(awsCreds, profileName, awsSharedCredentialsFile)
	return awsIdentity, nil
}

func (p *ServiceProvider) setProfileName(profileName string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue("static-profile") {
		p.logger.Debug("static profile name found")
		item := cfg.Get("static-profile")
		profileName = item.Value.(string)
	}
	p.logger.Debugw("setting aws profile name", "profile", profileName)

	item, err := cfg.String("aws-profile", profileName, "AWS profile name to use")
	if err != nil {
		return fmt.Errorf("setting aws-profile: %w", err)
	}
	item.Value = profileName

	return nil
}

func (p *ServiceProvider) resolveRole(awsRoles []*saml2aws.AWSRole, samlAssertion string, account *cfg.IDPAccount, roleFilter string) (*saml2aws.AWSRole, error) {
	if len(awsRoles) == 1 {
		if account.RoleARN != "" {
			return saml2aws.LocateRole(awsRoles, account.RoleARN)
		}
		return awsRoles[0], nil
	} else if len(awsRoles) == 0 {
		return nil, ErrNoRolesFound
	}

	// TODO: change this so its passed in
	samlAssertionData, err := base64.StdEncoding.DecodeString(samlAssertion)
	if err != nil {
		//TODO: change to specific error
		return nil, err
	}

	aud, err := p.extractDestinationURL(samlAssertionData)
	if err != nil {
		//TODO: return a better error
		return nil, fmt.Errorf("extracting destination url: %w", err)
	}

	awsAccounts, err := saml2aws.ParseAWSAccounts(aud, samlAssertion)
	if err != nil {
		//TODO: handle error better
		return nil, err
	}
	if len(awsAccounts) == 0 {
		return nil, ErrNotAccounts
	}

	saml2aws.AssignPrincipals(awsRoles, awsAccounts)

	awsAccounts = p.filterAccounts(awsAccounts, roleFilter)

	if account.RoleARN != "" {
		return saml2aws.LocateRole(awsRoles, account.RoleARN)
	}

	role, err := p.getRoleFromPrompt(awsAccounts, roleFilter)
	if err != nil {
		return nil, fmt.Errorf("getting role: %w", err)
	}

	return role, nil
}

func (p *ServiceProvider) filterAccounts(accounts []*saml2aws.AWSAccount, roleFilter string) []*saml2aws.AWSAccount {
	if roleFilter == "" {
		return accounts
	}

	filtered := []*saml2aws.AWSAccount{}
	for _, account := range accounts {
		filteredAccount := &saml2aws.AWSAccount{
			Name:  account.Name,
			Roles: []*saml2aws.AWSRole{},
		}
		for _, awsRole := range account.Roles {
			if strings.Contains(awsRole.RoleARN, roleFilter) {
				filteredAccount.Roles = append(filteredAccount.Roles, awsRole)
			}
		}
		if len(filteredAccount.Roles) > 0 {
			filtered = append(filtered, filteredAccount)
		}
	}

	return filtered
}

// Not using saml2aws.PromptForAWSRoleSelection as we want to implement custom logic
func (p *ServiceProvider) getRoleFromPrompt(accounts []*saml2aws.AWSAccount, roleFilter string) (*saml2aws.AWSRole, error) {

	roles := map[string]*saml2aws.AWSRole{}
	roleOptions := map[string]string{}

	for _, account := range accounts {
		for _, role := range account.Roles {
			if roleFilter == "" || strings.Contains(role.RoleARN, roleFilter) {
				name := fmt.Sprintf("%s / %s", account.Name, role.Name)
				roles[name] = role
				roleOptions[name] = name
			}
		}
	}

	selected, err := p.itemSelector("Select AWS role", roleOptions)
	if err != nil {
		return nil, fmt.Errorf("selected aws role: %w", err)
	}
	p.logger.Debugw("selected aws role", "name", selected, "arn", roles[selected].RoleARN)

	return roles[selected], nil
}

func (p *ServiceProvider) loginToStsUsingRole(account *cfg.IDPAccount, role *saml2aws.AWSRole, samlAssertion string) (*awsconfig.AWSCredentials, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: &account.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("creating aws session: %w", err)
	}

	svc := sts.New(sess)

	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(role.PrincipalARN),
		RoleArn:         aws.String(role.RoleARN),
		SAMLAssertion:   aws.String(samlAssertion),
		DurationSeconds: aws.Int64(int64(account.SessionDuration)),
	}

	p.logger.Info("requesting AWS credentials using SAML")

	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return nil, fmt.Errorf("retrieving STS credentials using SAML: %w", err)
	}

	return &awsconfig.AWSCredentials{
		AWSAccessKey:     aws.StringValue(resp.Credentials.AccessKeyId),
		AWSSecretKey:     aws.StringValue(resp.Credentials.SecretAccessKey),
		AWSSessionToken:  aws.StringValue(resp.Credentials.SessionToken),
		AWSSecurityToken: aws.StringValue(resp.Credentials.SessionToken),
		PrincipalARN:     aws.StringValue(resp.AssumedRoleUser.Arn),
		Expires:          resp.Credentials.Expiration.Local(),
		Region:           account.Region,
	}, nil
}

// TODO: use the version form saml2aws when modules are fixed
func (p *ServiceProvider) extractDestinationURL(data []byte) (string, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return "", err
	}

	rootElement := doc.Root()
	if rootElement == nil {
		return "", fmt.Errorf("getting response element %s: %w", responseTag, ErrMissingResponseElement)
	}

	destination := rootElement.SelectAttrValue("Destination", "none")
	if destination != "none" {
		return destination, nil

	}

	confirmData := doc.FindElement("//SubjectConfirmationData")
	if confirmData != nil {
		recipient := confirmData.SelectAttr("Recipient")
		if recipient != nil {
			return recipient.Value, nil
		}
	}

	return "", fmt.Errorf("getting response element Destination or SubjectConfirmationData: %w", ErrMissingResponseElement)
}

func (p *ServiceProvider) Validate(configItems config.ConfigurationSet) error {
	cfg := &awsProviderConfig{}

	if err := config.Unmarshall(configItems, cfg); err != nil {
		return fmt.Errorf("unmarshlling config set: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validating config struct: %w", err)
	}

	return nil
}

func (p *ServiceProvider) ConfigurationItems() config.ConfigurationSet {
	cs := kaws.SharedConfig()

	return cs
}
