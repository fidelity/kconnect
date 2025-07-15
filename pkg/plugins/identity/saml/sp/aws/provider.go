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
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	saml2aws "github.com/versent/saml2aws/v2"
	"github.com/versent/saml2aws/v2/pkg/awsconfig"
	"github.com/versent/saml2aws/v2/pkg/cfg"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	kcfg "github.com/fidelity/kconnect/pkg/config"
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
	Region    string `json:"region"    validate:"required"`
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

func (p *ServiceProvider) ConfigurationItems() kcfg.ConfigurationSet {
	cs := kaws.SharedConfig()

	return cs
}

func (p *ServiceProvider) PopulateAccount(account *cfg.IDPAccount, cfg kcfg.ConfigurationSet) error {
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

func (p *ServiceProvider) ProcessAssertions(account *cfg.IDPAccount, samlAssertions string, cfg kcfg.ConfigurationSet) (identity.Identity, error) {
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

	// switch AWS IAM role
	assumeRoleARN := cfg.Get("assume-role-arn")
	if assumeRoleARN != nil && assumeRoleARN.Value.(string) != "" {
		awsCreds, err = p.assumeRoleARN(account, awsCreds, assumeRoleARN.Value.(string))
		if err != nil {
			return nil, fmt.Errorf("assuming role in AWS: %w", err)
		}

		if err := cfg.SetValue("assume-role-arn", assumeRoleARN.Value.(string)); err != nil {
			return nil, fmt.Errorf("setting assume-role-arn config value: %w", err)
		}

		p.logger.Debugw("role assumed", "assume-role", assumeRoleARN.Value.(string))
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

func (p *ServiceProvider) Validate(configItems kcfg.ConfigurationSet) error {
	cfg := &awsProviderConfig{}

	if err := kcfg.Unmarshall(configItems, cfg); err != nil {
		return fmt.Errorf("unmarshlling config set: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validating config struct: %w", err)
	}

	return nil
}

func (p *ServiceProvider) setProfileName(profileName string, cfg kcfg.ConfigurationSet) error {
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
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(account.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("creating aws session: %w", err)
	}

	svc := sts.NewFromConfig(cfg)

	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(role.PrincipalARN),
		RoleArn:         aws.String(role.RoleARN),
		SAMLAssertion:   aws.String(samlAssertion),
		DurationSeconds: aws.Int32(int32(account.SessionDuration)),
	}

	p.logger.Info("requesting AWS credentials using SAML")

	resp, err := svc.AssumeRoleWithSAML(context.TODO(), params)
	if err != nil {
		return nil, fmt.Errorf("retrieving STS credentials using SAML: %w", err)
	}

	return &awsconfig.AWSCredentials{
		AWSAccessKey:     aws.ToString(resp.Credentials.AccessKeyId),
		AWSSecretKey:     aws.ToString(resp.Credentials.SecretAccessKey),
		AWSSessionToken:  aws.ToString(resp.Credentials.SessionToken),
		AWSSecurityToken: aws.ToString(resp.Credentials.SessionToken),
		PrincipalARN:     aws.ToString(resp.AssumedRoleUser.Arn),
		Expires:          resp.Credentials.Expiration.Local(),
		Region:           account.Region,
	}, nil
}

func (p *ServiceProvider) assumeRoleARN(account *cfg.IDPAccount, awsCreds *awsconfig.AWSCredentials, assumeRoleARN string) (*awsconfig.AWSCredentials, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(account.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsCreds.AWSAccessKey,
			awsCreds.AWSSecretKey,
			awsCreds.AWSSessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating aws session: %w", err)
	}

	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         &assumeRoleARN,
		RoleSessionName: &account.Username,
		DurationSeconds: aws.Int32(int32(account.SessionDuration)),
	}

	out, err := sts.NewFromConfig(cfg).AssumeRole(context.TODO(), assumeRoleInput)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	return &awsconfig.AWSCredentials{
		AWSAccessKey:     aws.ToString(out.Credentials.AccessKeyId),
		AWSSecretKey:     aws.ToString(out.Credentials.SecretAccessKey),
		AWSSessionToken:  aws.ToString(out.Credentials.SessionToken),
		AWSSecurityToken: aws.ToString(out.Credentials.SessionToken),
		PrincipalARN:     aws.ToString(out.AssumedRoleUser.Arn),
		Expires:          out.Credentials.Expiration.Local(),
		Region:           account.Region,
	}, nil
}

// TODO: use the version from saml2aws when modules are fixed
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
