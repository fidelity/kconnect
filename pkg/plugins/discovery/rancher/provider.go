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

package rancher

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
	rshared "github.com/fidelity/kconnect/pkg/rancher"
)

const (
	ProviderName = "rancher"
)

func init() {
	if err := provider.RegisterClusterProviderPlugin(ProviderName, newProvider()); err != nil {
		zap.S().Fatalw("Failed to register Rancher cluster provider plugin", "error", err)
	}
}

func newProvider() *rancherClusterProvider {
	return &rancherClusterProvider{}
}

type rancherClusterProviderConfig struct {
	provider.ClusterProviderConfig
	rshared.CommonConfig
}

type rancherClusterProvider struct {
	config *rancherClusterProviderConfig
	token  string

	logger *zap.SugaredLogger
}

func (p *rancherClusterProvider) Name() string {
	return ProviderName
}

func (p *rancherClusterProvider) SupportedIDs() []string {
	return []string{"static-token", "rancher-ad"}
}

func (p *rancherClusterProvider) ConfigurationItems() config.ConfigurationSet {
	cs := config.NewConfigurationSet()
	rshared.AddCommonConfig(cs) //nolint: errcheck

	return cs
}

func (p *rancherClusterProvider) ConfigurationResolver() provider.ConfigResolver {
	return p
}

func (p *rancherClusterProvider) setup(cs config.ConfigurationSet, identity provider.Identity) error {
	p.ensureLogger()
	cfg := &rancherClusterProviderConfig{}
	if err := config.Unmarshall(cs, cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into rancherClusterProviderConfig: %w", err)
	}
	p.config = cfg

	id, ok := identity.(*provider.TokenIdentity)
	if !ok {
		return provider.ErrNotTokenIdentity
	}
	p.token = id.Token()

	return nil
}

func (p *rancherClusterProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", ProviderName)
	}
}

// UsageExample will provide an example of the usage of this provider
func (p *rancherClusterProvider) UsageExample() string {
	return `  # Discover Rancher clusters using Active Directory
  {{.CommandPath}} use rancher --idp-protocol rancher-ad

  # Discover clusters via Rancher using a API key
  {{.CommandPath}} use rancher --idp-protocol static-token --token ABCDEF
`
}

func (p *rancherClusterProvider) CheckPreReqs() error {
	return nil
}
