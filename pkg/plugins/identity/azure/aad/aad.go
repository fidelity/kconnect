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

package aad

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/plugins/discovery/azure"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	provid "github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
)

const (
	ProviderName = "aad"
)

var (
	ErrAddingCommonCfg = errors.New("adding common identity config")
	ErrNoExpiryTime    = errors.New("token has no expiry time")
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
		zap.S().Fatalw("Failed to register Azure Active Directory identity plugin", "error", err)
	}
}

// New will create a new saml identity provider
func New(input *provider.PluginCreationInput) (provid.Provider, error) {
	if input.HTTPClient == nil {
		return nil, provider.ErrHTTPClientRequired
	}

	return &aadIdentityProvider{
		logger:      input.Logger,
		interactive: input.IsInteractice,
		httpClient:  input.HTTPClient,
	}, nil
}

type aadIdentityProvider struct {
	interactive bool
	logger      *zap.SugaredLogger
	httpClient  khttp.Client
}

type aadConfig struct {
	common.IdentityProviderConfig

	TenantID string           `json:"tenant-id"`
	ClientID string           `json:"client-id"`
	AADHost  identity.AADHost `json:"aad-host"`
}

func (p *aadIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and return details of their identity.
func (p *aadIdentityProvider) Authenticate(ctx context.Context, input *provid.AuthenticateInput) (*provid.AuthenticateOutput, error) {
	p.logger.Info("authenticating user")

	cfg := &aadConfig{}
	if err := config.Unmarshall(input.ConfigSet, cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into use aadconfig: %w", err)
	}

	authCfg := &identity.AuthenticationConfig{
		Authority: &identity.AuthorityConfig{
			Tenant:       cfg.TenantID,
			Host:         cfg.AADHost,
			AuthorityURI: fmt.Sprintf("https://%s/%s/", cfg.AADHost, cfg.TenantID),
		},
		ClientID: cfg.ClientID,
		Username: cfg.Username,
		Password: cfg.Password,
	}

	endpointResolver := identity.NewOAuthEndpointsResolver(p.httpClient)
	endpoints, err := endpointResolver.Resolve(authCfg.Authority)
	if err != nil {
		return nil, fmt.Errorf("getting endpoints: %w", err)
	}
	authCfg.Endpoints = endpoints

	identityClient := identity.NewClient(p.httpClient)
	userRealm, err := identityClient.GetUserRealm(authCfg)
	if err != nil {
		return nil, fmt.Errorf("getting user realm: %w", err)
	}

	id := identity.NewActiveDirectoryIdentity(authCfg, userRealm, ProviderName, p.httpClient)

	return &provid.AuthenticateOutput{
		Identity: id,
	}, nil
}

func (p *aadIdentityProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *aadIdentityProvider) CheckPreReqs() error {
	return nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()

	if err := common.AddCommonIdentityConfig(cs); err != nil {
		return nil, ErrAddingCommonCfg
	}

	cs.String(azure.TenantIDConfigItem, "", "The azure tenant id")                                        //nolint: errcheck
	cs.String(azure.ClientIDConfigItem, "04b07795-8ddb-461a-bbee-02f9e1bf7b46", "The azure ad client id") //nolint: errcheck
	cs.String(azure.AADHostConfigItem, string(identity.AADHostWorldwide), "The AAD host to use")          //nolint: errcheck

	cs.SetShort(azure.TenantIDConfigItem, "t") //nolint: errcheck
	cs.SetRequired(azure.TenantIDConfigItem)   //nolint: errcheck
	cs.SetRequired(azure.ClientIDConfigItem)   //nolint: errcheck
	cs.SetRequired(azure.AADHostConfigItem)    //nolint: errcheck

	return cs, nil
}
