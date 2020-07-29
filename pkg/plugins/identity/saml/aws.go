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

package saml

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/beevik/etree"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/provider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/versent/saml2aws"
	"github.com/versent/saml2aws/pkg/awsconfig"
	"github.com/versent/saml2aws/pkg/cfg"
)

var (
	ErrUnexpectedIdentity = errors.New("unexpected identity type")
	ErrNoRegion           = errors.New("no region found")
	ErrNoProfile          = errors.New("no profile found")
	ErrNoRolesFound       = errors.New("no aws roles found")
	ErrNotAccounts        = errors.New("no accounts available")
)

func newAWSIdentityStore(providerFlags *pflag.FlagSet) (provider.IdentityStore, error) {
	if !flags.ExistsWithValue("profile", providerFlags) {
		return nil, ErrNoProfile
	}
	profileFlag := providerFlags.Lookup("profile")

	return &awsIdentityStore{
		configProvider: awsconfig.NewSharedCredentials(profileFlag.Value.String()),
	}, nil
}

type awsIdentityStore struct {
	configProvider *awsconfig.CredentialsProvider
}

func (s *awsIdentityStore) CredsExists() (bool, error) {
	return s.configProvider.CredsExists()
}

func (s *awsIdentityStore) Save(identity provider.Identity) error {
	awsIdentity, ok := identity.(*AWSIdentity)
	if !ok {
		return fmt.Errorf("expected AWSIdentity but got a %T: %w", identity, ErrUnexpectedIdentity)
	}
	awsCreds := mapIdentityToCreds(awsIdentity)

	return s.configProvider.Save(awsCreds)
}

func (s *awsIdentityStore) Load() (provider.Identity, error) {
	creds, err := s.configProvider.Load()
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}
	awsID := mapCredsToIdentity(creds, s.configProvider.Profile)

	return awsID, nil
}

func (s *awsIdentityStore) Expired() bool {
	return s.configProvider.Expired()
}

type AWSIdentity struct {
	ProfileName      string
	AWSAccessKey     string
	AWSSecretKey     string
	AWSSessionToken  string
	AWSSecurityToken string
	PrincipalARN     string
	Expires          time.Time
	Region           string
}

func newAWSIdentity(profileName string) *AWSIdentity {
	return &AWSIdentity{
		ProfileName: profileName,
	}
}

type awsServiveProvider struct {
	logger *log.Entry
}

func (p *awsServiveProvider) PopulateAccount(account *cfg.IDPAccount, flags *pflag.FlagSet) error {
	account.AmazonWebservicesURN = "urn:amazon:webservices"

	regionFlag := flags.Lookup("region")
	if regionFlag == nil || regionFlag.Value.String() == "" {
		return ErrNoRegion
	}
	account.Region = regionFlag.Value.String()

	profileFlag := flags.Lookup("profile")
	if profileFlag == nil || profileFlag.Value.String() == "" {
		return ErrNoProfile
	}
	account.Profile = profileFlag.Value.String()

	roleFlag := flags.Lookup("role-arn")
	if roleFlag != nil || roleFlag.Value.String() != "" {
		account.RoleARN = roleFlag.Value.String()
	}
	account.Region = regionFlag.Value.String()

	return nil
}

func (p *awsServiveProvider) ProcessAssertions(account *cfg.IDPAccount, samlAssertions string) (provider.Identity, error) {
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

func (p *awsServiveProvider) resolveRole(awsRoles []*saml2aws.AWSRole, samlAssertion string, account *cfg.IDPAccount) (*saml2aws.AWSRole, error) {
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
		//TODO: change tpo specific error
		return nil, err
	}

	aud, err := p.extractDestinationURL(samlAssertionData)
	if err != nil {
		//TODO: return a better error
		return nil, fmt.Errorf("extracting destination utl: %w", err)
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

func (p *awsServiveProvider) loginToStsUsingRole(account *cfg.IDPAccount, role *saml2aws.AWSRole, samlAssertion string) (*awsconfig.AWSCredentials, error) {
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
func (p *awsServiveProvider) extractDestinationURL(data []byte) (string, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return "", err
	}

	rootElement := doc.Root()
	if rootElement == nil {
		return "", fmt.Errorf("getting response element %s: %w", responseTag, ErrMissingResponseElement)
	}

	destination := rootElement.SelectAttrValue("Destination", "none")
	if destination == "none" {
		return "", fmt.Errorf("getting response element Destination: %w", ErrMissingResponseElement)
	}

	return destination, nil
}

func (p *awsServiveProvider) Resolver() provider.FlagsResolver {
	return NewAWSFlagsResolver(p.logger)
}

func mapCredsToIdentity(creds *awsconfig.AWSCredentials, profileName string) *AWSIdentity {
	return &AWSIdentity{
		AWSAccessKey:     creds.AWSAccessKey,
		AWSSecretKey:     creds.AWSSecretKey,
		AWSSecurityToken: creds.AWSSecurityToken,
		AWSSessionToken:  creds.AWSSessionToken,
		Expires:          creds.Expires,
		PrincipalARN:     creds.PrincipalARN,
		ProfileName:      profileName,
	}
}

func mapIdentityToCreds(awsIdentity *AWSIdentity) *awsconfig.AWSCredentials {
	return &awsconfig.AWSCredentials{
		AWSAccessKey:     awsIdentity.AWSAccessKey,
		AWSSecretKey:     awsIdentity.AWSSecretKey,
		AWSSecurityToken: awsIdentity.AWSSecurityToken,
		AWSSessionToken:  awsIdentity.AWSSessionToken,
		Expires:          awsIdentity.Expires,
		PrincipalARN:     awsIdentity.PrincipalARN,
	}
}
