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
	"fmt"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider"
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin("aad", newProvider()); err != nil {
		zap.S().Fatalw("failed to register Azure AD identity provider plugin", "error", err)
	}
}

func newProvider() *aadIdentityProvider {
	return &aadIdentityProvider{}
}

type aadIdentityProvider struct {
	logger *zap.SugaredLogger
}

type aadConfig struct {
	provider.IdentityProviderConfig

	TenantID string           `json:"tenant-id"`
	ClientID string           `json:"client-id"`
	AADHost  identity.AADHost `json:"aad-host"`
}

func (p *aadIdentityProvider) Name() string {
	return "aad"
}

// Authenticate will authenticate a user and return details of
// their identity.
func (p *aadIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	p.ensureLogger()
	p.logger.Info("authenticating user")

	cfg := &aadConfig{}
	if err := config.Unmarshall(ctx.ConfigurationItems(), cfg); err != nil {
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

	httpClient := khttp.NewHTTPClient()
	endpointResolver := identity.NewOIDCEndpointsResolver(httpClient)
	endpoints, err := endpointResolver.Resolve(authCfg.Authority)
	if err != nil {
		return nil, fmt.Errorf("getting endpoints: %w", err)
	}
	authCfg.Endpoints = endpoints

	identityClient := identity.NewClient(httpClient)
	userRealm, err := identityClient.GetUserRealm(authCfg)
	if err != nil {
		return nil, fmt.Errorf("getting user realm: %w", err)
	}

	switch userRealm.AccountType {
	case identity.AccountTypeFederated:
		_, _ = identityClient.GetMex(userRealm.FederationMetadataURL)
	case identity.AccountTypeManaged:
	default:
		return nil, identity.ErrUnknownAccountType
	}

	return nil, nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func (p *aadIdentityProvider) ConfigurationItems(clusterProviderName string) (config.ConfigurationSet, error) {
	p.ensureLogger()
	cs := config.NewConfigurationSet()

	if err := provider.AddCommonIdentityConfig(cs); err != nil {
		return nil, fmt.Errorf("adding common identity config")
	}

	cs.String("tenant-id", "", "the azure tenant id")
	cs.String("client-id", "04b07795-8ddb-461a-bbee-02f9e1bf7b46", "the azure ad client id")
	cs.String("aad-host", string(identity.AADHostWorldwide), "The AAD host to use")

	cs.SetShort("tenant-id", "t")
	cs.SetRequired("tenant-id")

	return cs, nil
}

// Usage returns a string to display for help
func (p *aadIdentityProvider) Usage(clusterProvider string) (string, error) {
	return "", nil
}

func (p *aadIdentityProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", "aad")
	}
}
