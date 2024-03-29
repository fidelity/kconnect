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
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
	"github.com/fidelity/kconnect/pkg/utils"
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
	p.logParameters()

	return nil
}

func (p *oidcClusterProvider) logParameters() {

	p.logger.Infof("Using oidc-server-url: %s.", p.identity.OidcServer)
	p.logger.Infof("Using oidc-client-id: %s.", p.identity.OidcId)
	if p.identity.UsePkce == True {
		p.logger.Infof("using pkce.")
	}
	p.logger.Infof("Using cluster-id: %s.", *p.config.ClusterID)
	p.logger.Infof("Using cluster-url: %s.", p.config.ClusterUrl)
	p.logger.Infof("Using cluster-auth: %s.", p.config.ClusterAuth)

}

// For required parameter, if not exists in default config file or config url, read user input.
func (p *oidcClusterProvider) readRequiredFields() error {

	if p.config.ClusterUrl == "" {
		value, err := oidc.ReadUserInput(oidc.ClusterUrlConfigItem, oidc.ClusterUrlConfigDescription)
		if err != nil {
			return err
		}
		p.config.ClusterUrl = value
	}

	if p.config.ClusterAuth == "" {
		value, err := oidc.ReadUserInput(oidc.ClusterAuthConfigItem, oidc.ClusterAuthConfigDescription)
		if err != nil {
			return err
		}
		p.config.ClusterAuth = value
	}

	if *p.config.ClusterID == "" {
		value, err := oidc.ReadUserInput(oidc.ClusterIdConfigItem, oidc.ClusterIdConfigDescription)
		if err != nil {
			return err
		}
		p.config.ClusterID = &value
	}

	return nil

}

func (p *oidcClusterProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *oidcClusterProvider) CheckPreReqs() error {
	return utils.CheckKubectlOidcLoginPrereq()
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
