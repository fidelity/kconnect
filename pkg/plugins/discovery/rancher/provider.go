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
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
	rshared "github.com/fidelity/kconnect/pkg/rancher"
)

const (
	ProviderName = "rancher"
	UsageExample = `  # Discover Rancher clusters using Active Directory
	{{.CommandPath}} use rancher --idp-protocol rancher-ad

	# Discover clusters via Rancher using a API key
	{{.CommandPath}} use rancher --idp-protocol static-token --token ABCDEF
  `
)

func init() {
	if err := registry.RegisterDiscoveryPlugin(&registry.DiscoveryPluginRegistration{
		PluginRegistration: registry.PluginRegistration{
			Name:                   ProviderName,
			UsageExample:           UsageExample,
			ConfigurationItemsFunc: ConfigurationItems,
		},
		CreateFunc:                 New,
		SupportedIdentityProviders: []string{"static-token", "rancher-ad"},
	}); err != nil {
		zap.S().Fatalw("Failed to register Rancher discovery plugin", "error", err)
	}
}

// New will create a new Rancher discovery plugin
func New(input *provider.PluginCreationInput) (discovery.Provider, error) {
	if input.HTTPClient == nil {
		return nil, provider.ErrHTTPClientRequired
	}

	return &rancherClusterProvider{
		logger:      input.Logger,
		interactive: input.IsInteractice,
		httpClient:  input.HTTPClient,
	}, nil
}

type rancherClusterProviderConfig struct {
	common.ClusterProviderConfig
	rshared.CommonConfig
}

type rancherClusterProvider struct {
	config *rancherClusterProviderConfig
	token  string

	httpClient  khttp.Client
	interactive bool
	logger      *zap.SugaredLogger
}

func (p *rancherClusterProvider) Name() string {
	return ProviderName
}

func (p *rancherClusterProvider) setup(cs config.ConfigurationSet, userID provider.Identity) error {
	cfg := &rancherClusterProviderConfig{}
	if err := config.Unmarshall(cs, cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into rancherClusterProviderConfig: %w", err)
	}
	p.config = cfg

	id, ok := userID.(*identity.TokenIdentity)
	if !ok {
		return identity.ErrNotTokenIdentity
	}
	p.token = id.Token()

	return nil
}

func (p *rancherClusterProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *rancherClusterProvider) CheckPreReqs() error {
	return nil
}

func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()
	rshared.AddCommonConfig(cs) //nolint: errcheck

	return cs, nil
}
