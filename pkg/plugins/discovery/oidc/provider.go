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

package oidc

import (
	"context"
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/oidc"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
	"go.uber.org/zap"
)

const (
	ProviderName = "oidc"
	UsageExample = `
  # Setup cluster for oidc protocol using default config file
  {{.CommandPath}} use oidc

  # Setup cluster for oidc protocol using config url
  {{.CommandPath}} use oidc --config-url https://localhost:8080

  # Setup cluster and add an alias to its connection history entry
  {{.CommandPath}} use oidc --alias mycluster
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
		SupportedIdentityProviders: []string{"oidc"},
	}); err != nil {
		zap.S().Fatalw("Failed to register OIDC discovery plugin", "error", err)
	}
}

// New will create a new OIDC discovery plugin
func New(input *provider.PluginCreationInput) (discovery.Provider, error) {
	return &oidcClusterProvider{
		logger: input.Logger,
	}, nil
}

type oidcClusterProviderConfig struct {
	common.ClusterProviderConfig
	ClusterUrl  string `json:"cluster-url"`
	ClusterAuth string `json:"cluster-auth"`
	ConfigUrl   string `json:"config-url"`
	ConfigCA    string `json:"ca-cert"`
}

// oidcClusterProvider will generate OIDC config
type oidcClusterProvider struct {
	config   *oidcClusterProviderConfig
	identity *oidc.Identity
	logger   *zap.SugaredLogger
}

// Name returns the name of the provider
func (p *oidcClusterProvider) Name() string {
	return ProviderName
}

func (p *oidcClusterProvider) setup(cs config.ConfigurationSet, userID identity.Identity) error {
	cfg := &oidcClusterProviderConfig{}
	if err := config.Unmarshall(cs, cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into oidcClusterProviderConfig: %w", err)
	}
	p.config = cfg

	oidcID, ok := userID.(*oidc.Identity)
	if !ok {
		return fmt.Errorf("unmarshalling config items into oidcClusterProviderConfig")
	}
	p.identity = oidcID

	if err := p.readRequiredFields(); err != nil {
		return err
	}

	return nil
}

// For required parameter, if not exists in default config file or config url, read user input.
func (p *oidcClusterProvider) readRequiredFields() error {

	if p.identity.OidcId == "" {
		value, err := readUserInput(oidc.OidcIdConfigItem, oidc.OidcIdConfigDescription)
		if err != nil {
			return err
		}
		p.identity.OidcId = value
	}

	if p.identity.OidcServer == "" {
		value, err := readUserInput(oidc.OidcServerConfigItem, oidc.OidcServerConfigDescription)
		if err != nil {
			return err
		}
		p.identity.OidcServer = value
	}

	if p.identity.UsePkce == "" && p.identity.OidcSecret == "" {
		value, err := readUserInput(oidc.OidcSecretConfigItem, oidc.OidcSecretConfigDescription)
		if err != nil {
			return err
		}
		p.identity.OidcSecret = value
	}

	if p.config.ClusterUrl == "" {
		value, err := readUserInput(oidc.ClusterUrlConfigItem, oidc.ClusterUrlConfigDescription)
		if err != nil {
			return err
		}
		p.config.ClusterUrl = value
	}

	if p.config.ClusterAuth == "" {
		value, err := readUserInput(oidc.ClusterAuthConfigItem, oidc.ClusterAuthConfigDescription)
		if err != nil {
			return err
		}
		p.config.ClusterAuth = value
	}

	if *p.config.ClusterID == "" {
		value, err := readUserInput(oidc.ClusterIdConfigItem, oidc.ClusterIdConfigDescription)
		if err != nil {
			return err
		}
		p.config.ClusterID = &value
	}

	return nil

}

func readUserInput(key string, msg string) (string, error) {
	userInput, err := prompt.Input(key, msg, false)
	if userInput == "" || err != nil {
		return userInput, fmt.Errorf("error reading input for %s", key)
	}
	return userInput, nil
}

func (p *oidcClusterProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *oidcClusterProvider) CheckPreReqs() error {
	return nil
}

// ConfigurationItems returns the configuration items for this provider
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	return oidc.SharedConfig(), nil
}

func (p *oidcClusterProvider) getCluster(ctx context.Context) (*discovery.Cluster, error) {
	cluster := discovery.Cluster{
		ID:                       *p.config.ClusterID,
		Name:                     *p.config.ClusterID,
		ControlPlaneEndpoint:     &p.config.ClusterUrl,
		CertificateAuthorityData: &p.config.ClusterAuth,
	}
	return &cluster, nil
}
