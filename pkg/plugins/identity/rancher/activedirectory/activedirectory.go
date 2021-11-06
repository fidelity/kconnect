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

package activedirectory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/defaults"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
	"github.com/fidelity/kconnect/pkg/rancher"
)

const (
	ProviderName = "rancher-ad"
)

var (
	ErrAddingCommonCfg      = errors.New("adding common identity config")
	ErrNoExpiryTime         = errors.New("token has no expiry time")
	ErrAuthenticationFailed = errors.New("failed to authenticate using active directory")
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
		zap.S().Fatalw("Failed to register Rancher Active Directory identity plugin", "error", err)
	}
}

// New will create a new Rancher Active Directory identity provider
func New(input *provider.PluginCreationInput) (identity.Provider, error) {
	if input.HTTPClient == nil {
		return nil, provider.ErrHTTPClientRequired
	}

	return &radIdentityProvider{
		logger:      input.Logger,
		interactive: input.IsInteractice,
		httpClient:  input.HTTPClient,
	}, nil
}

type radIdentityProvider struct {
	interactive bool
	logger      *zap.SugaredLogger
	httpClient  khttp.Client
}

type radConfig struct {
	common.IdentityProviderConfig
	rancher.CommonConfig
}

func (p *radIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and return details of
// their identity.
func (p *radIdentityProvider) Authenticate(ctx context.Context, input *identity.AuthenticateInput) (*identity.AuthenticateOutput, error) {
	p.logger.Info("authenticating user")

	cfg := &radConfig{}
	if err := config.Unmarshall(input.ConfigSet, cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into use radConfig: %w", err)
	}

	resolver, err := rancher.NewStaticEndpointsResolver(cfg.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("vreating endpoint resolver: %w", err)
	}

	loginRequest := &loginRequest{
		Type:        "token",
		Description: "automation",
		Username:    cfg.Username,
		Password:    cfg.Password,
	}

	data, err := json.Marshal(loginRequest)
	if err != nil {
		return nil, fmt.Errorf("marshalling ad request: %w", err)
	}

	headers := defaults.Headers(defaults.WithNoCache(), defaults.WithContentTypeJSON())

	resp, err := p.httpClient.Post(resolver.ActiveDirectoryAuth(), string(data), headers)
	if err != nil {
		return nil, fmt.Errorf("performing AD auth: %w", err)
	}

	if resp.ResponseCode() != http.StatusCreated {
		return nil, ErrAuthenticationFailed
	}

	loginResponse := &loginResponse{}
	if err := json.Unmarshal([]byte(resp.Body()), loginResponse); err != nil {
		return nil, fmt.Errorf("unmarshalling login response: %w", err)
	}

	id := identity.NewTokenIdentity(loginResponse.UserID, loginResponse.Token, ProviderName)

	return &identity.AuthenticateOutput{
		Identity: id,
	}, nil
}

func (p *radIdentityProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *radIdentityProvider) CheckPreReqs() error {
	return nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()

	if err := common.AddCommonIdentityConfig(cs); err != nil {
		return nil, ErrAddingCommonCfg
	}
	if err := rancher.AddCommonConfig(cs); err != nil {
		return nil, ErrAddingCommonCfg
	}

	return cs, nil
}
