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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider"
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
	if err := provider.RegisterIdentityProviderPlugin(ProviderName, newProvider()); err != nil {
		zap.S().Fatalw("failed to register Rancher AD identity provider plugin", "error", err)
	}
}

func newProvider() *radIdentityProvider {
	return &radIdentityProvider{}
}

type radIdentityProvider struct {
	logger *zap.SugaredLogger
}

type radConfig struct {
	provider.IdentityProviderConfig
	rancher.CommonConfig
}

func (p *radIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and return details of
// their identity.
func (p *radIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	p.ensureLogger()
	p.logger.Info("authenticating user")

	if err := p.resolveConfig(ctx); err != nil {
		return nil, fmt.Errorf("resolving config: %w", err)
	}

	cfg := &radConfig{}
	if err := config.Unmarshall(ctx.ConfigurationItems(), cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into use radConfig: %w", err)
	}

	if err := p.validateConfig(cfg); err != nil {
		return nil, err
	}

	resolver, err := rancher.NewStaticEndpointsResolver(cfg.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("vreating endpoint resolver: %w", err)
	}

	input := &loginRequest{
		Type:        "token",
		Description: "automation",
		Username:    cfg.Username,
		Password:    cfg.Password,
	}

	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshalling ad request: %w", err)
	}

	headers := defaults.Headers(defaults.WithNoCache(), defaults.WithContentTypeJSON())
	httpClient := khttp.NewHTTPClient()

	resp, err := httpClient.Post(resolver.ActiveDirectoryAuth(), string(data), headers)
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

	id := provider.NewTokenIdentity(loginResponse.UserID, loginResponse.Token, ProviderName)

	return id, nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func (p *radIdentityProvider) ConfigurationItems(clusterProviderName string) (config.ConfigurationSet, error) {
	p.ensureLogger()
	cs := config.NewConfigurationSet()

	if err := provider.AddCommonIdentityConfig(cs); err != nil {
		return nil, ErrAddingCommonCfg
	}
	if err := rancher.AddCommonConfig(cs); err != nil {
		return nil, ErrAddingCommonCfg
	}

	return cs, nil
}

// Usage returns a string to display for help
func (p *radIdentityProvider) Usage(clusterProvider string) (string, error) {
	return "", nil
}

func (p *radIdentityProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", ProviderName)
	}
}

func (p *radIdentityProvider) validateConfig(cfg *radConfig) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validating aad config: %w", err)
	}
	return nil
}
