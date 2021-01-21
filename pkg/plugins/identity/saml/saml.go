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
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/versent/saml2aws"
	"github.com/versent/saml2aws/pkg/cfg"
	"github.com/versent/saml2aws/pkg/creds"
	"go.uber.org/zap"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
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
	ProviderName   = "saml"
	defaultSession = 3600
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin(ProviderName, newSAMLProvider()); err != nil {
		zap.S().Fatalw("failed to register SAML identity provider plugin", "error", err)
	}
}

func newSAMLProvider() *samlIdentityProvider {
	return &samlIdentityProvider{
		itemSelector: provider.DefaultItemSelection,
	}
}

type samlIdentityProvider struct {
	config          *sp.ProviderConfig
	serviceProvider sp.ServiceProvider
	itemSelector    provider.SelectItemFunc

	logger *zap.SugaredLogger
}

// Name returns the name of the plugin
func (p *samlIdentityProvider) Name() string {
	return ProviderName
}

func (p *samlIdentityProvider) ConfigurationItems(clusterProviderName string) (config.ConfigurationSet, error) {
	p.ensureLogger()
	cs := config.NewConfigurationSet()

	cs.String("idp-endpoint", "", "identity provider endpoint provided by your IT team") //nolint: errcheck
	cs.String("idp-provider", "", "the name of the idp provider")                        //nolint: errcheck
	cs.SetRequired("idp-endpoint")                                                       //nolint: errcheck
	cs.SetRequired("idp-provider")                                                       //nolint: errcheck

	// get the service provider flags
	if clusterProviderName != "" {
		sp, err := p.createServiceProvider(clusterProviderName)
		if err != nil {
			return nil, fmt.Errorf("creating saml service provider: %w", err)
		}

		spConfig := sp.ConfigurationItems()
		if err := cs.AddSet(spConfig); err != nil {
			return nil, fmt.Errorf("adding service provider config: %w", err)
		}
	}

	return cs, nil
}

// Authenticate will authenticate a user and returns their identity
func (p *samlIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	p.ensureLogger()
	p.logger.Info("authenticating user")

	sp, err := p.createServiceProvider(clusterProvider)
	if err != nil {
		return nil, fmt.Errorf("creating saml service provider: %w", err)
	}
	p.serviceProvider = sp

	if err := p.resolveConfig(ctx); err != nil {
		return nil, err
	}

	if err := p.bindAndValidateConfig(ctx.ConfigurationItems()); err != nil {
		return nil, fmt.Errorf("binding and validation config: %w", err)
	}
	if err := p.serviceProvider.Validate(ctx.ConfigurationItems()); err != nil {
		return nil, fmt.Errorf("validating service provider: %w", err)
	}

	account, err := p.createAccount(ctx.ConfigurationItems())
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

	userID, err := p.serviceProvider.ProcessAssertions(account, samlAssertion, ctx.ConfigurationItems())
	if err != nil {
		return nil, fmt.Errorf("processing assertions for: %s: %w", clusterProvider, err)
	}

	store, err := p.createIdentityStore(ctx, clusterProvider)
	if err != nil {
		return nil, fmt.Errorf("creating identity store for %s: %w", clusterProvider, err)
	}

	err = store.Save(userID)
	if err != nil {
		return nil, fmt.Errorf("saving identity: %w", err)
	}

	return userID, nil
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

func (p *samlIdentityProvider) createServiceProvider(clusterProviderName string) (sp.ServiceProvider, error) {
	switch clusterProviderName {
	case "eks":
		return aws.NewServiceProvider(p.itemSelector), nil
	default:
		return nil, ErrUnsuportedProvider
	}
}

func (p *samlIdentityProvider) resolveConfig(ctx *provider.Context) error {
	sp := p.serviceProvider

	if ctx.IsInteractive() {
		p.logger.Debug("resolving SAML provider flags")
		if err := sp.ResolveConfiguration(ctx.ConfigurationItems()); err != nil {
			return fmt.Errorf("resolving flags: %w", err)
		}
	} else {
		p.logger.Debug("skipping configuration resolution as runnning non-interactive")
	}

	return nil
}

func (p *samlIdentityProvider) createIdentityStore(ctx *provider.Context, clusterProviderName string) (provider.IdentityStore, error) {
	var store provider.IdentityStore
	var err error

	switch clusterProviderName {
	case "eks":
		if !ctx.ConfigurationItems().ExistsWithValue("aws-profile") {
			return nil, kaws.ErrNoProfile
		}
		profileCfg := ctx.ConfigurationItems().Get("aws-profile")
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

// Usage returns the usage for the provider
func (p *samlIdentityProvider) Usage(clusterProvider string) (string, error) {
	p.ensureLogger()
	usage := []string{"SAML idp-protocol Flags:"}

	cfg, err := p.ConfigurationItems(clusterProvider)
	if err != nil {
		return "", fmt.Errorf("getting provider configuration: %w", err)
	}

	fs, err := flags.CreateFlagsFromConfig(cfg)
	if err != nil {
		return "", err
	}

	usage = append(usage, fs.FlagUsages())

	return strings.Join(usage, "\n"), nil
}

func (p *samlIdentityProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", ProviderName)
	}
}
