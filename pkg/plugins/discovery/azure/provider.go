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

	"github.com/fidelity/kconnect/pkg/config"
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
	TenantID       *string `json:"tenant-id"`
	SubscriptionID *string `json:"subscription-id"`
	Region         *string `json:"region"`
	RegionFilter   *string `json:"region-filter"`
}

type aksClusterProvider struct {
	config *aksClusterProviderConfig

	logger *zap.SugaredLogger
}

func (p *aksClusterProvider) Name() string {
	return providerName
}

// ConfigurationItems returns the configuration items for this provider
func (p *aksClusterProvider) ConfigurationItems() config.ConfigurationSet {
	cs := config.NewConfigurationSet()

	cs.String("region-filter", "", "A filter to apply to the AWS regions list, e.g. 'us-' will only show US regions") //nolint: errcheck
	cs.String("tenant-id", "", "The Azure tenant to use")
	cs.String("subscription-id", "", "The Azure subscription to use")

	return cs
}

// ConfigurationResolver returns the resolver to use for config with this provider
func (p *aksClusterProvider) ConfigurationResolver() provider.ConfigResolver {
	return &azureConfigResolver{}
}

func (p *aksClusterProvider) setup(ctx *provider.Context, identity provider.Identity) error {
	p.ensureLogger()
	cfg := &aksClusterProviderConfig{}
	if err := config.Unmarshall(ctx.ConfigurationItems(), cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into eksClusteProviderConfig: %w", err)
	}
	p.config = cfg

	//TODO: get the identity and create the client

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
