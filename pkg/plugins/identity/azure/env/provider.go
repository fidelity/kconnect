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

package env

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	"github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	ProviderName = "az-env"
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin(ProviderName, newProvider()); err != nil {
		zap.S().Fatalw("failed to register azure env identity provider plugin", "error", err)
	}
}

func newProvider() *envIdentityProvider {
	return &envIdentityProvider{}
}

type envIdentityProvider struct {
	logger *zap.SugaredLogger
}

type providerConfig struct {
	UseFile bool `json:"use-file"`
}

func (p *envIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and return details of their envIdentityProvider.
func (p *envIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	p.ensureLogger()
	p.logger.Info("using azure environment for authentication")

	cfg := &providerConfig{}
	if err := config.Unmarshall(ctx.ConfigurationItems(), cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into providerConfig: %w", err)
	}

	var authorizer autorest.Authorizer
	var err error
	if cfg.UseFile {
		authorizer, err = auth.NewAuthorizerFromFile("")
	} else {
		authorizer, err = auth.NewAuthorizerFromEnvironment()
	}

	if err != nil {
		return nil, fmt.Errorf("getting authorizer: %w", err)
	}

	id := identity.NewAuthorizerIdentity("", authorizer)

	return id, nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func (p *envIdentityProvider) ConfigurationItems(clusterProviderName string) (config.ConfigurationSet, error) {
	p.ensureLogger()
	cs := config.NewConfigurationSet()

	cs.Bool("use-file", false, "Use file based authorization") //nolint:errcheck

	return cs, nil
}

// Usage returns a string to display for help
func (p *envIdentityProvider) Usage(clusterProvider string) (string, error) {
	return "", nil
}

func (p *envIdentityProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", ProviderName)
	}
}
