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
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/versent/saml2aws"
	"github.com/versent/saml2aws/pkg/cfg"
	"github.com/versent/saml2aws/pkg/creds"
	"go.uber.org/zap"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp"
	"github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp/aws"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
)

var (
	ErrNoClusterProvider  = errors.New("no cluster provider on the context")
	ErrUnsuportedProvider = errors.New("cluster provider not supported")
	ErrNoSAMLAssertions   = errors.New("no SAML assertions")
	ErrCreatingAccount    = errors.New("creating account")
)

const (
	ProviderName   = "saml"
	defaultSession = 3600
)

var (
	ErrRequiresScope = errors.New("saml provider required to be scoped")
)

func init() {
	if err := registry.RegisterIdentityPlugin(&registry.IdentityPluginRegistration{
		PluginRegistration: registry.PluginRegistration{
			Name:                   ProviderName,
			UsageExample:           "",
			ConfigurationItemsFunc: ConfigurationItems,
		},
		CreateFunc: New,
	}); err != nil {
		zap.S().Fatalw("Failed to register SAML identity plugin", "error", err)
	}
}

// New will create a new saml identity provider
func New(input *provider.PluginCreationInput) (identity.Provider, error) {
	if input.ScopedTo == nil || *input.ScopedTo == "" {
		return nil, ErrRequiresScope
	}

	return &samlIdentityProvider{
		logger:            input.Logger,
		interactive:       input.IsInteractice,
		itemSelector:      input.ItemSelector,
		scopedToDiscovery: *input.ScopedTo,
	}, nil
}

type samlIdentityProvider struct {
	config            *sp.ProviderConfig
	serviceProvider   sp.ServiceProvider
	scopedToDiscovery string

	itemSelector provider.SelectItemFunc
	interactive  bool
	logger       *zap.SugaredLogger
}

// Name returns the name of the plugin
func (p *samlIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and returns their identity
func (p *samlIdentityProvider) Authenticate(ctx context.Context, input *identity.AuthenticateInput) (*identity.AuthenticateOutput, error) {
	p.logger.Info("authenticating user")

	sp, err := createServiceProvider(p.scopedToDiscovery, p.itemSelector)
	if err != nil {
		return nil, fmt.Errorf("creating saml service provider: %w", err)
	}
	p.serviceProvider = sp

	if err := p.resolveConfig(input.ConfigSet); err != nil {
		return nil, err
	}

	if err := p.bindAndValidateConfig(input.ConfigSet); err != nil {
		return nil, fmt.Errorf("binding and validation config: %w", err)
	}
	if err := p.serviceProvider.Validate(input.ConfigSet); err != nil {
		return nil, fmt.Errorf("validating service provider: %w", err)
	}

	account, err := p.createAccount(input.ConfigSet)
	if err != nil {
		return nil, ErrCreatingAccount
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

	userID, err := p.serviceProvider.ProcessAssertions(account, samlAssertion, input.ConfigSet)
	if err != nil {
		return nil, fmt.Errorf("processing assertions for: %s: %w", p.scopedToDiscovery, err)
	}

	store, err := p.createIdentityStore(input.ConfigSet)
	if err != nil {
		return nil, fmt.Errorf("creating identity store for %s: %w", p.scopedToDiscovery, err)
	}

	err = store.Save(userID)
	if err != nil {
		return nil, fmt.Errorf("saving identity: %w", err)
	}

	return &identity.AuthenticateOutput{
		Identity: userID,
	}, nil
}

func (p *samlIdentityProvider) bindAndValidateConfig(cs config.ConfigurationSet) error {
	spConfig := &sp.ProviderConfig{}

	if err := config.Unmarshall(cs, spConfig); err != nil {
		return fmt.Errorf("unmarshalling configuration: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(spConfig); err != nil {
		return fmt.Errorf("validating config struct: %w", err)
	}

	p.config = spConfig

	return nil
}

func (p *samlIdentityProvider) createAccount(cs config.ConfigurationSet) (*cfg.IDPAccount, error) {
	account := &cfg.IDPAccount{
		URL:             p.config.IdpEndpoint,
		Provider:        p.config.IdpProvider,
		MFA:             "Auto",
		SessionDuration: defaultSession,
	}
	if err := p.serviceProvider.PopulateAccount(account, cs); err != nil {
		return nil, fmt.Errorf("populating account: %w", err)
	}

	return account, nil
}

func (p *samlIdentityProvider) resolveConfig(cfg config.ConfigurationSet) error {
	sp := p.serviceProvider

	if p.interactive {
		p.logger.Debug("resolving SAML provider flags")
		if err := sp.ResolveConfiguration(cfg); err != nil {
			return fmt.Errorf("resolving flags: %w", err)
		}
	} else {
		p.logger.Debug("skipping configuration resolution as runnning non-interactive")
	}

	return nil
}

func (p *samlIdentityProvider) createIdentityStore(cfg config.ConfigurationSet) (identity.Store, error) {
	var store identity.Store
	var err error

	switch p.scopedToDiscovery {
	case "eks":
		if !cfg.ExistsWithValue("aws-profile") {
			return nil, kaws.ErrNoProfile
		}
		profileCfg := cfg.Get("aws-profile")
		profile := profileCfg.Value.(string)
		store, err = kaws.NewIdentityStore(profile, ProviderName)
	default:
		return nil, ErrUnsuportedProvider
	}
	if err != nil {
		return nil, fmt.Errorf("creating identity store: %w", err)
	}

	return store, nil
}

func createServiceProvider(clusterProviderName string, itemSelector provider.SelectItemFunc) (sp.ServiceProvider, error) {
	switch clusterProviderName {
	case "eks":
		return aws.NewServiceProvider(itemSelector), nil
	default:
		return nil, ErrUnsuportedProvider
	}
}

func ConfigurationItems(scopedToDiscovery string) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()

	cs.String("idp-endpoint", "", "identity provider endpoint provided by your IT team") //nolint: errcheck
	cs.String("idp-provider", "", "the name of the idp provider")                        //nolint: errcheck
	cs.SetRequired("idp-endpoint")                                                       //nolint: errcheck
	cs.SetRequired("idp-provider")                                                       //nolint: errcheck

	// get the service provider flags
	sp, err := createServiceProvider(scopedToDiscovery, nil)
	if err != nil {
		return nil, fmt.Errorf("creating saml service provider: %w", err)
	}

	spConfig := sp.ConfigurationItems()
	if err := cs.AddSet(spConfig); err != nil {
		return nil, fmt.Errorf("adding service provider config: %w", err)
	}

	return cs, nil
}
