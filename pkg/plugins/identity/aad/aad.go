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
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/oidc"
	"github.com/fidelity/kconnect/pkg/provider"
)

var (
	ErrAddingCommonCfg = errors.New("adding common identity config")
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

	var token *identity.OauthToken
	switch userRealm.AccountType {
	case identity.AccountTypeFederated:
		doc, err := identityClient.GetMex(userRealm.FederationMetadataURL)
		if err != nil {
			return nil, err //TODO: specific error
		}
		endpoint := doc.UsernamePasswordEndpoint
		wsTrustResponse, err := identityClient.GetWsTrustResponse(authCfg, userRealm.CloudAudienceURN, &endpoint)
		if err != nil {
			return nil, err //TODO: specific error
		}
		token, err = identityClient.GetOauth2TokenFromSamlAssertion(authCfg, wsTrustResponse.Body.RequestSecurityTokenResponseCollection.RequestSecurityTokenResponse.RequestedSecurityToken.Assertion, "https://management.azure.com/")
		if err != nil {
			return nil, err //TODO: specific error
		}

	case identity.AccountTypeManaged:
		token, err = identityClient.GetOauth2TokenFromUsernamePassword(authCfg)
		if err != nil {
			return nil, err //TODO: specific error
		}
	case identity.AccountTypeUnknown: //TODO: need to check other parts
		token, err = identityClient.GetOauth2TokenFromUsernamePassword(authCfg)
		if err != nil {
			return nil, err //TODO: specific error
		}
	default:
		return nil, identity.ErrUnknownAccountType
	}

	id, err := p.createIdentityFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("creating identity from oauth token: %w", err)
	}

	return id, nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func (p *aadIdentityProvider) ConfigurationItems(clusterProviderName string) (config.ConfigurationSet, error) {
	p.ensureLogger()
	cs := config.NewConfigurationSet()

	if err := provider.AddCommonIdentityConfig(cs); err != nil {
		return nil, ErrAddingCommonCfg
	}

	cs.String("tenant-id", "", "The azure tenant id")                                        //nolint: errcheck
	cs.String("client-id", "04b07795-8ddb-461a-bbee-02f9e1bf7b46", "The azure ad client id") //nolint: errcheck
	cs.String("aad-host", string(identity.AADHostWorldwide), "The AAD host to use")          //nolint: errcheck

	cs.SetShort("tenant-id", "t") //nolint: errcheck
	cs.SetRequired("tenant-id")   //nolint: errcheck

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

func (p *aadIdentityProvider) createIdentityFromToken(token *identity.OauthToken) (*oidc.Identity, error) {
	id := &oidc.Identity{
		AccessToken:  token.AccessToken,
		IDToken:      token.IDToken,
		RefreshToken: token.RefreshToken,
		Resource:     token.Resource,
		Scope:        token.Scope,
	}

	expiresSeconds, err := token.ExpiresOn.Int64()
	if err != nil {
		return nil, fmt.Errorf("getting ExpiresOn: %w", err)
	}
	id.Expires = time.Unix(expiresSeconds, 0)

	return id, nil
}
