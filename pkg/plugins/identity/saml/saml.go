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
	"github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp"
	"github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp/aws"
	"github.com/fidelity/kconnect/pkg/provider"
)

var (
	ErrNoClusterProvider  = errors.New("no cluster provider on the context")
	ErrUnsuportedProvider = errors.New("cluster provider not supported")
	ErrNoSAMLAssertions   = errors.New("no SAML assertions")
	ErrCreatingAccount    = errors.New("creating account")
)

const (
	defaultSession = 3600
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
	config          *sp.ProviderConfig
	logger          *logrus.Entry
	serviceProvider sp.ServiceProvider
	store           provider.IdentityStore
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
	p.logger = ctx.Logger().WithField("provider", "saml")
	p.logger.Info("Authenticating user")

	if err := p.createServiceProvider(clusterProvider); err != nil {
		return nil, fmt.Errorf("setting up saml provider: %w", err)
	}

	if err := p.resolveAndValidateFlags(ctx); err != nil {
		return nil, err
	}

	if err := p.bindFlags(ctx); err != nil {
		return nil, fmt.Errorf("binding flags: %w", err)
	}

	if err := p.createStore(ctx, clusterProvider); err != nil {
		return nil, fmt.Errorf("creating identity store: %w", err)
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
			p.logger.Info("using cached creds")
			id, err := p.store.Load()
			if err != nil {
				return nil, fmt.Errorf("loading idebtity: %w", err)
			}
			return id, nil
		}
		p.logger.Info("cached creds expired, renewing")
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
		Username: p.config.Username,
		Password: p.config.Password,
		URL:      p.config.IdpEndpoint,
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

func (p *samlIdentityProvider) bindFlags(ctx *provider.Context) error {
	cfg := &sp.ProviderConfig{}
	if err := flags.Unmarshal(ctx.Command().Flags(), cfg); err != nil {
		return fmt.Errorf("unmarshalling flags into config: %w", err)
	}
	p.config = cfg

	return nil
}

func (p *samlIdentityProvider) createStore(ctx *provider.Context, providerName string) error {
	store, err := p.createIdentityStore(ctx, providerName)
	if err != nil {
		return fmt.Errorf("creating identity store for %s: %w", providerName, err)
	}
	p.store = store

	return nil
}

func (p *samlIdentityProvider) createAccount(ctx *provider.Context) (*cfg.IDPAccount, error) {
	account := &cfg.IDPAccount{
		URL:             p.config.IdpEndpoint,
		Provider:        p.config.IdpProvider,
		MFA:             "Auto",
		SessionDuration: defaultSession,
	}
	if err := p.serviceProvider.PopulateAccount(account, ctx.Command().Flags()); err != nil {
		return nil, fmt.Errorf("populating account: %w", err)
	}

	return account, nil
}

func (p *samlIdentityProvider) resolveAndValidateFlags(ctx *provider.Context) error {
	sp := p.serviceProvider

	if ctx.IsInteractive() {
		p.logger.Debug("running interactively, resolving SAML provider flags")
		if err := sp.ResolveFlags(ctx, ctx.Command().Flags()); err != nil {
			return fmt.Errorf("resolving flags: %w", err)
		}
	}

	if err := sp.Validate(ctx, ctx.Command().Flags()); err != nil {
		return fmt.Errorf("validating supplied flags: %w", err)
	}

	return nil
}

func (p *samlIdentityProvider) createServiceProvider(providerName string) error {
	switch providerName {
	case "eks":
		p.serviceProvider = aws.NewServiceProvider(p.logger)
	default:
		return ErrUnsuportedProvider
	}

	return nil
}

func (p *samlIdentityProvider) createIdentityStore(ctx *provider.Context, providerName string) (provider.IdentityStore, error) {
	var store provider.IdentityStore
	var err error

	switch providerName {
	case "eks":
		store, err = aws.NewIdentityStore(ctx.Command().Flags())
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
