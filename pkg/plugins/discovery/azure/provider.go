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

package azure

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/Azure/go-autorest/autorest"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/oidc"
	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	providerName = "aks"
)

func init() {
	if err := provider.RegisterClusterProviderPlugin(providerName, newAKSProvider()); err != nil {
		zap.S().Fatalw("Failed to register AKS cluster provider plugin", "error", err)
	}
}

func newAKSProvider() *aksClusterProvider {
	return &aksClusterProvider{}
}

type aksClusterProviderConfig struct {
	provider.ClusterProviderConfig
	SubscriptionID   *string `json:"subscription-id"`
	SubscriptionName *string `json:"subscription-name"`
	ResourceGroup    *string `json:"resource-group"`
	Admin            bool    `json:"admin"`
}

type aksClusterProvider struct {
	config     *aksClusterProviderConfig
	authorizer *autorest.BearerAuthorizer

	logger *zap.SugaredLogger
}

func (p *aksClusterProvider) Name() string {
	return providerName
}

func (p *aksClusterProvider) SupportedIDs() []string {
	return []string{"aad"}
}

// ConfigurationItems returns the configuration items for this provider
func (p *aksClusterProvider) ConfigurationItems() config.ConfigurationSet {
	cs := config.NewConfigurationSet()

	cs.String(SubscriptionIDConfigItem, "", "The Azure subscription to use (specified by ID)")     //nolint: errcheck
	cs.String(SubscriptionNameConfigItem, "", "The Azure subscription to use (specified by name)") //nolint: errcheck
	cs.String(ResourceGroupConfigItem, "", "The Azure resource group to use")                      //nolint: errcheck
	cs.Bool(AdminConfigItem, false, "Generate admin user kubeconfig")                              //nolint: errcheck

	cs.SetShort(ResourceGroupConfigItem, "r") //nolint: errcheck

	return cs
}

// ConfigurationResolver returns the resolver to use for config with this provider
func (p *aksClusterProvider) ConfigurationResolver() provider.ConfigResolver {
	return p
}

func (p *aksClusterProvider) setup(cs config.ConfigurationSet, identity provider.Identity) error {
	p.ensureLogger()
	cfg := &aksClusterProviderConfig{}
	if err := config.Unmarshall(cs, cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into eksClusteProviderConfig: %w", err)
	}
	p.config = cfg

	id, ok := identity.(*oidc.Identity)
	if !ok {
		return ErrNotOIDCIdentity
	}

	p.logger.Debugw("creating bearer authorizer")
	bearerAuth := autorest.NewBearerAuthorizer(id)
	p.authorizer = bearerAuth

	return nil
}

func (p *aksClusterProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", providerName)
	}
}

// UsageExample will provide an example of the usage of this provider
func (p *aksClusterProvider) UsageExample() string {
	return `  # Discover AKS clusters using Azure AD
  kconnect use aks --idp-protocol aad
`
}
