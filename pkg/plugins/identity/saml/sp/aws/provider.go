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

	"github.com/beevik/etree"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/versent/saml2aws"
	"github.com/versent/saml2aws/pkg/awsconfig"
	"github.com/versent/saml2aws/pkg/cfg"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp"
	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	responseTag = "Response"
)

var (
	ErrUnexpectedIdentity     = errors.New("unexpected identity type")
	ErrNoRegion               = errors.New("no region found")
	ErrNoProfile              = errors.New("no profile supplied")
	ErrNoRolesFound           = errors.New("no aws roles found")
	ErrNotAccounts            = errors.New("no accounts available")
	ErrMissingResponseElement = errors.New("missing response element")
	ErrNoPartitionSupplied    = errors.New("no AWS partition supplied")
	ErrPartitionNotFound      = errors.New("AWS partition not found")
)

type awsProviderConfig struct {
	sp.ProviderConfig

	Partition string `json:"partition" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Profile   string `json:"profile" validate:"required"`
}

func NewServiceProvider(logger *log.Entry) sp.ServiceProvider {
	return &ServiceProvider{
		logger: logger.WithField("serviceprovider", "aws"),
	}
}

type ServiceProvider struct {
	logger *log.Entry
}

func (p *ServiceProvider) PopulateAccount(account *cfg.IDPAccount, cfg config.ConfigurationSet) error {
	account.AmazonWebservicesURN = "urn:amazon:webservices"

	regionCfg := cfg.Get("region")
	if regionCfg == nil || regionCfg.Value.(string) == "" {
		return ErrNoRegion
	}
	account.Region = regionCfg.Value.(string)

	profileCfg := cfg.Get("profile")
	if profileCfg == nil || profileCfg.Value.(string) == "" {
		return ErrNoProfile
	}
	account.Profile = profileCfg.Value.(string)

	roleCfg := cfg.Get("role-arn")
	if roleCfg != nil || roleCfg.Value.(string) != "" {
		account.RoleARN = roleCfg.Value.(string)
	}

	return nil
}

func (p *ServiceProvider) ProcessAssertions(account *cfg.IDPAccount, samlAssertions string) (provider.Identity, error) {
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

	role, err := p.resolveRole(awsRoles, samlAssertions, account)
	if err != nil {
		return nil, fmt.Errorf("resolving aws role: %w", err)
	}

	log.Printf("selected role: %s", role.RoleARN)

	awsCreds, err := p.loginToStsUsingRole(account, role, samlAssertions)
	if err != nil {
		return nil, fmt.Errorf("logging into AWS using STS and SAMLAssertion: %w", err)
	}

	awsIdentity := mapCredsToIdentity(awsCreds, account.Profile)
	return awsIdentity, nil
}

func (p *ServiceProvider) resolveRole(awsRoles []*saml2aws.AWSRole, samlAssertion string, account *cfg.IDPAccount) (*saml2aws.AWSRole, error) {
	var role = new(saml2aws.AWSRole)

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

	if account.RoleARN != "" {
		return saml2aws.LocateRole(awsRoles, account.RoleARN)
	}

	for {
		role, err = saml2aws.PromptForAWSRoleSelection(awsAccounts)
		if err == nil {
			break
		}
		log.Println("Error selecting role, try again")
	}

	return role, nil
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

	log.Println("Requesting AWS credentials using SAML")

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
	//TODO: handle this
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

func mapCredsToIdentity(creds *awsconfig.AWSCredentials, profileName string) *Identity {
	return &Identity{
		AWSAccessKey:     creds.AWSAccessKey,
		AWSSecretKey:     creds.AWSSecretKey,
		AWSSecurityToken: creds.AWSSecurityToken,
		AWSSessionToken:  creds.AWSSessionToken,
		Expires:          creds.Expires,
		PrincipalARN:     creds.PrincipalARN,
		ProfileName:      profileName,
	}
}

func mapIdentityToCreds(awsIdentity *Identity) *awsconfig.AWSCredentials {
	return &awsconfig.AWSCredentials{
		AWSAccessKey:     awsIdentity.AWSAccessKey,
		AWSSecretKey:     awsIdentity.AWSSecretKey,
		AWSSecurityToken: awsIdentity.AWSSecurityToken,
		AWSSessionToken:  awsIdentity.AWSSessionToken,
		Expires:          awsIdentity.Expires,
		PrincipalARN:     awsIdentity.PrincipalARN,
	}
}
