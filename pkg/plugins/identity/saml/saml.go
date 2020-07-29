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
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/versent/saml2aws"
	"github.com/versent/saml2aws/pkg/cfg"
	"github.com/versent/saml2aws/pkg/creds"

	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	responseTag    = "Response"
	defaultSession = 3600
)

var (
	ErrNoClusterProvider      = errors.New("no cluster provider on the context")
	ErrUnsuportedProvider     = errors.New("cluster provider not supported")
	ErrMissingResponseElement = errors.New("missing response element")
	ErrNoSAMLAssertions       = errors.New("no SAML assertions")
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin("saml", newSAMLProvider()); err != nil {
		// TODO: handle fatal error
		logrus.Fatalf("Failed to register SAML identity provider plugin: %v", err)
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

type samlServiceProvider interface {
	Resolver() provider.FlagsResolver
	PopulateAccount(account *cfg.IDPAccount, flags *pflag.FlagSet) error
	ProcessAssertions(account *cfg.IDPAccount, samlAssertions string) (provider.Identity, error)
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

		//TODO: how to handle flags applicable to all providers
		p.username = p.flags.String("username", "", "the username used for authentication")
		p.password = p.flags.String("password", "", "the password to use for authentication")

	}

	return p.flags
}

// Authenticate will authenticate a user and returns their identity
func (p *samlIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	logger := ctx.Logger().WithField("provider", "saml")
	logger.Info("Authenticating user")

	sp, err := p.createServiceProvider(ctx, clusterProvider, logger)
	if err != nil {
		return nil, fmt.Errorf("creating service provider for %s: %w", clusterProvider, err)
	}
	if ctx.IsInteractive() {
		err = sp.Resolver().Resolve(ctx, ctx.Command().Flags())
		if err != nil {
			return nil, fmt.Errorf("resolving flags: %w", err)
		}
	}
	err = sp.Resolver().Validate(ctx, ctx.Command().Flags())
	if err != nil {
		return nil, fmt.Errorf("validating supplied flags: %w", err)
	}

	store, err := p.createIdentityStore(ctx, clusterProvider)
	if err != nil {
		return nil, fmt.Errorf("creating identity store for %s: %w", clusterProvider, err)
	}

	account := &cfg.IDPAccount{
		URL:             *p.idpEndpoint,
		Provider:        *p.idpProvider,
		MFA:             "Auto",
		SessionDuration: defaultSession,
	}
	err = sp.PopulateAccount(account, ctx.Command().Flags())
	if err != nil {
		return nil, fmt.Errorf("populating account: %w", err)
	}

	exist, err := store.CredsExists()
	if err != nil {
		return nil, fmt.Errorf("checking if creds exist: %w", err)
	}
	if exist {
		if !store.Expired() {
			logger.Info("using cached creds")
			return newAWSIdentity(account.Profile), nil
		}
		logger.Info("cached creds expired, renewing")
	}

	err = account.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating saml: %w", err)
	}

	client, err := saml2aws.NewSAMLClient(account)
	if err != nil {
		return nil, fmt.Errorf("creating saml client: %w", err)
	}

	loginDetails := &creds.LoginDetails{
		Username: *p.username,
		Password: *p.password,
		URL:      *p.idpEndpoint,
	}

	samlAssertion, err := client.Authenticate(loginDetails)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	if samlAssertion == "" {
		return nil, ErrNoSAMLAssertions
	}

	userID, err := sp.ProcessAssertions(account, samlAssertion)
	if err != nil {
		return nil, fmt.Errorf("processing assertions for: %s: %w", clusterProvider, err)
	}

	err = store.Save(userID)
	if err != nil {
		return nil, fmt.Errorf("saving identity: %w", err)
	}

	return userID, nil
}

func (p *samlIdentityProvider) createServiceProvider(_ *provider.Context, providerName string, logger *logrus.Entry) (samlServiceProvider, error) {
	switch providerName {
	case "eks":
		return &awsServiveProvider{logger: logger}, nil
	default:
		return nil, ErrUnsuportedProvider
	}
}

func (p *samlIdentityProvider) createIdentityStore(ctx *provider.Context, providerName string) (provider.IdentityStore, error) {
	var store provider.IdentityStore
	var err error

	switch providerName {
	case "eks":
		store, err = newAWSIdentityStore(ctx.Command().Flags())
	default:
		return nil, ErrUnsuportedProvider
	}
	if err != nil {
		return nil, fmt.Errorf("creating identity store: %w", err)
	}

	return store, nil
}

// Usage returns a description for use in the help/usage
func (p *samlIdentityProvider) Usage() string {
	return "SAML Idp authentication"
}
