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

package token

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	ProviderName    = "static-token"
	tokenConfigItem = "token"
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin(ProviderName, newProvider()); err != nil {
		zap.S().Fatalw("failed to register static token identity provider plugin", "error", err)
	}
}

func newProvider() *staticTokenIdentityProvider {
	return &staticTokenIdentityProvider{}
}

type staticTokenIdentityProvider struct {
	logger *zap.SugaredLogger
}

type providerConfig struct {
	Token string `json:"token" validate:"required"`
}

func (p *staticTokenIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and return details of
// their identity.
func (p *staticTokenIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	p.ensureLogger()
	p.logger.Info("using static token for authentication")

	if err := p.resolveConfig(ctx); err != nil {
		return nil, fmt.Errorf("resolving config: %w", err)
	}

	cfg := &providerConfig{}
	if err := config.Unmarshall(ctx.ConfigurationItems(), cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into providerConfig: %w", err)
	}

	if err := p.validateConfig(cfg); err != nil {
		return nil, err
	}

	id := provider.NewTokenIdentity("static-token", cfg.Token, ProviderName)

	return id, nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func (p *staticTokenIdentityProvider) ConfigurationItems(clusterProviderName string) (config.ConfigurationSet, error) {
	p.ensureLogger()
	cs := config.NewConfigurationSet()

	cs.String("idp-protocol", "", "The idp protocol to use (e.g. saml). Each protocol has its own flags.") //nolint:errcheck
	cs.String(tokenConfigItem, "", "the token to use for authentication")                                  //nolint:errcheck
	cs.SetRequired(tokenConfigItem)                                                                        //nolint:errcheck
	cs.SetSensitive(tokenConfigItem)                                                                       //nolint:errcheck

	return cs, nil
}

// Usage returns a string to display for help
func (p *staticTokenIdentityProvider) Usage(clusterProvider string) (string, error) {
	return "", nil
}

func (p *staticTokenIdentityProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", ProviderName)
	}
}

func (p *staticTokenIdentityProvider) validateConfig(cfg *providerConfig) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validating aad config: %w", err)
	}
	return nil
}

func (p *staticTokenIdentityProvider) resolveConfig(ctx *provider.Context) error {
	if !ctx.IsInteractive() {
		p.logger.Debug("skipping configuration resolution as runnning non-interactive")
	}

	cfg := ctx.ConfigurationItems()

	if err := prompt.Input(cfg, tokenConfigItem, "Enter authentication token", true); err != nil {
		return fmt.Errorf("resolving %s: %w", tokenConfigItem, err)
	}

	return nil
}
