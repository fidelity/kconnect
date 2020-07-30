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

	"github.com/fidelity/kconnect/pkg/flags"
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
	ErrCreatingAccount        = errors.New("creating account")
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin("saml", newSAMLProvider()); err != nil {
		// TODO: handle fatal error
		logrus.Fatalf("Failed to register SAML identity provider plugin: %v", err)
	}
}

type samlProviderConfig struct {
	provider.IdentityProviderConfig
	IdpEndpoint *string `flag:"idp-endpoint"`
	IdpProvider *string `flag:"idp-provider"`
}

func newSAMLProvider() *samlIdentityProvider {
	return &samlIdentityProvider{}
}

type samlIdentityProvider struct {
	config          *samlProviderConfig
	logger          *logrus.Entry
	serviceProvider samlServiceProvider
	store           provider.IdentityStore
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
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.String("idp-endpoint", "", "identity provider endpoint provided by your IT team")
	fs.String("idp-provider", "", "the name of the idp provider")

	return fs
}

// Authenticate will authenticate a user and returns their identity
func (p *samlIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	logger := ctx.Logger().WithField("provider", "saml")
	logger.Info("Authenticating user")

	//TODO: should this be on the interface and the registry call it???
	if err := p.setup(ctx, clusterProvider, logger); err != nil {
		return nil, fmt.Errorf("setting up saml provider: %w", err)
	}

	if err := p.resolveAndValidateFlags(ctx); err != nil {
		return nil, err
	}

	account, err := p.createAccount(ctx)
	if err != nil {
		return nil, ErrCreatingAccount
	}

	exist, err := p.store.CredsExists()
	if err != nil {
		return nil, fmt.Errorf("checking if creds exist: %w", err)
	}
	if exist {
		if !p.store.Expired() {
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
		Username: *p.config.Username,
		Password: *p.config.Password,
		URL:      *p.config.IdpEndpoint,
	}

	samlAssertion, err := client.Authenticate(loginDetails)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	if samlAssertion == "" {
		return nil, ErrNoSAMLAssertions
	}

	userID, err := p.serviceProvider.ProcessAssertions(account, samlAssertion)
	if err != nil {
		return nil, fmt.Errorf("processing assertions for: %s: %w", clusterProvider, err)
	}

	err = p.store.Save(userID)
	if err != nil {
		return nil, fmt.Errorf("saving identity: %w", err)
	}

	return userID, nil
}

func (p *samlIdentityProvider) setup(ctx *provider.Context, providerName string, logger *logrus.Entry) error {
	cfg := &samlProviderConfig{}
	if err := flags.Unmarshal(ctx.Command().Flags(), cfg); err != nil {
		return fmt.Errorf("unmarshalling flags into config: %w", err)
	}

	sp, err := p.createServiceProvider(ctx, providerName, logger)
	if err != nil {
		return fmt.Errorf("creating SAML service provider for %s: %w", providerName, err)
	}
	p.serviceProvider = sp

	store, err := p.createIdentityStore(ctx, providerName)
	if err != nil {
		return fmt.Errorf("creating identity store for %s: %w", providerName, err)
	}
	p.store = store

	p.config = cfg
	p.logger = logger

	return nil
}

func (p *samlIdentityProvider) createAccount(ctx *provider.Context) (*cfg.IDPAccount, error) {
	account := &cfg.IDPAccount{
		URL:             *p.config.IdpEndpoint,
		Provider:        *p.config.IdpProvider,
		MFA:             "Auto",
		SessionDuration: defaultSession,
	}
	if err := p.serviceProvider.PopulateAccount(account, ctx.Command().Flags()); err != nil {
		return nil, fmt.Errorf("populating account: %w", err)
	}

	return account, nil
}

func (p *samlIdentityProvider) resolveAndValidateFlags(ctx *provider.Context) error {
	resolver := p.serviceProvider.Resolver()

	if ctx.IsInteractive() {
		p.logger.Debug("running interactively, resolving SAML provider flags")
		if err := resolver.Resolve(ctx, ctx.Command().Flags()); err != nil {
			return fmt.Errorf("resolving flags: %w", err)
		}
	}

	if err := resolver.Validate(ctx, ctx.Command().Flags()); err != nil {
		return fmt.Errorf("validating supplied flags: %w", err)
	}

	return nil
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
