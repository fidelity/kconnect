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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/beevik/etree"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/versent/saml2aws"
	"github.com/versent/saml2aws/pkg/awsconfig"
	"github.com/versent/saml2aws/pkg/cfg"
	"github.com/versent/saml2aws/pkg/creds"

	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	responseTag = "Response"
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin("saml", newSAMLProvider()); err != nil {
		// TODO: handle fatal error
		log.Fatalf("Failed to register SAML identity provider plugin: %w", err)
	}
}

func newSAMLProvider() *samlIdentityProvider {
	return &samlIdentityProvider{}
}

type samlIdentityProvider struct {
	idpEndpoint *string
	idpProvider *string
	username    *string
	password    *string

	flags *pflag.FlagSet
}

type AWSIdentity struct {
	profileName string
}

func NewAWSIdentity(profileName string) *AWSIdentity {
	return &AWSIdentity{
		profileName: profileName,
	}
}

func (i *AWSIdentity) Profile() string {
	return i.profileName
}

// Name returns the name of the plugin
func (p *samlIdentityProvider) Name() string {
	return "saml"
}

// Flags will return the flags for this plugin
func (p *samlIdentityProvider) Flags() *pflag.FlagSet {
	if p.flags == nil {
		p.flags = &pflag.FlagSet{}
		p.idpEndpoint = p.flags.String("idp-endpoint", "", "identity provider endpoint provided by your IT team")
		p.idpProvider = p.flags.String("idp-provider", "", "the name of the idp provider")

		p.username = p.flags.String("username", "", "the username used for authentication")
		p.password = p.flags.String("password", "", "the password to use for authentication")
	}

	return p.flags
}

// Authenticate will authenticate a user and returns their identity
func (p *samlIdentityProvider) Authenticate() (provider.Identity, error) {
	// cfmgr, err := cfg.NewConfigManager("")
	// if err != nil {
	// 	return nil, fmt.Errorf("getting config manager: %w", err)
	// }

	// account, err := cfmgr.LoadIDPAccount(*p.idpProvider)
	// if err != nil {
	// 	return nil, fmt.Errorf("getting idp account: %w", err)
	// }

	account := &cfg.IDPAccount{
		URL:                  *p.idpEndpoint,
		Provider:             "GoogleApps",
		MFA:                  "Auto",
		AmazonWebservicesURN: "urn:amazon:webservices",
		Profile:              "saml3",
		RoleARN:              "arn:aws:iam::482649550366:role/AdministratorAccess",
		Region:               "eu-west-2",
		SessionDuration:      3600,
	}

	sharedCreds := awsconfig.NewSharedCredentials(account.Profile)
	exist, err := sharedCreds.CredsExists()
	if err != nil {
		return nil, fmt.Errorf("checking if creds exist: %w", err)
	}
	if exist {
		if !sharedCreds.Expired() {
			log.Println("using cached creds")
			return NewAWSIdentity(account.Profile), nil
		} else {
			log.Println("cached creds expired, renewing")
		}
	}

	err = account.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating saml: %w", err)
	}

	client, err := saml2aws.NewSAMLClient(account)

	loginDetails := &creds.LoginDetails{
		Username: *p.username,
		Password: *p.password,
		URL:      *p.idpEndpoint,
	}

	samlAssertion, err := client.Authenticate(loginDetails)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	fmt.Println(samlAssertion)

	if samlAssertion == "" {
		//TODO: proper error here
		return nil, errors.New("no SAML assertaions")
	}

	data, err := base64.StdEncoding.DecodeString(samlAssertion)
	if err != nil {
		return nil, fmt.Errorf("decoding SAMLAssertion: %w", err)
	}

	// if provider == EKS
	roles, err := saml2aws.ExtractAwsRoles(data)
	if err != nil {
		return nil, fmt.Errorf("extracting AWS roles from assertion: %w", err)
	}

	if len(roles) == 0 {
		//TODO: handl this better
		return nil, nil
	}

	awsRoles, err := saml2aws.ParseAWSRoles(roles)
	if err != nil {
		return nil, fmt.Errorf("parsing aws roles: %w", err)
	}

	fmt.Println(awsRoles)

	role, err := p.resolveRole(awsRoles, samlAssertion, account)
	if err != nil {
		return nil, fmt.Errorf("resolving aws role: %w", err)
	}

	log.Printf("selected role: %s", role.RoleARN)

	awsCreds, err := loginToStsUsingRole(account, role, samlAssertion)
	if err != nil {
		return nil, fmt.Errorf("logging into AWS using STS and SAMLAssertion: %w", err)
	}

	// TODO: save the creds
	err = sharedCreds.Save(awsCreds)
	if err != nil {
		return nil, fmt.Errorf("saving aws credentials: %w", err)
	}

	return NewAWSIdentity(account.Profile), nil
}

func (p *samlIdentityProvider) resolveRole(awsRoles []*saml2aws.AWSRole, samlAssertion string, account *cfg.IDPAccount) (*saml2aws.AWSRole, error) {
	var role = new(saml2aws.AWSRole)

	if len(awsRoles) == 1 {
		if account.RoleARN != "" {
			return saml2aws.LocateRole(awsRoles, account.RoleARN)
		}
		return awsRoles[0], nil
	} else if len(awsRoles) == 0 {
		return nil, errors.New("no aws roles")
	}

	// TODO: change this so its passed in
	samlAssertionData, err := base64.StdEncoding.DecodeString(samlAssertion)
	if err != nil {
		//TODO: change tpo specific error
		return nil, err
	}

	aud, err := extractDestinationURL(samlAssertionData)
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
		return nil, errors.New("no accounts available")
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

// Usage returns a description for use in the help/usage
func (p *samlIdentityProvider) Usage() string {
	return "SAML Idp authentication"
}

func loginToStsUsingRole(account *cfg.IDPAccount, role *saml2aws.AWSRole, samlAssertion string) (*awsconfig.AWSCredentials, error) {
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
func extractDestinationURL(data []byte) (string, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return "", err
	}

	rootElement := doc.Root()
	if rootElement == nil {
		return "", fmt.Errorf("missing element: %s", responseTag)
	}

	destination := rootElement.SelectAttrValue("Destination", "none")
	if destination == "none" {
		return "", fmt.Errorf("missing element: %s", responseTag)
	}

	return destination, nil
}
