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
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/oidc"
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

	TenantID string           `json:"tenant-id" validate:"required"`
	ClientID string           `json:"client-id" validate:"required"`
	AADHost  identity.AADHost `json:"aad-host" validate:"required"`
}

func (p *aadIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and return details of their identity.
func (p *aadIdentityProvider) Authenticate(ctx context.Context, input *provid.AuthenticateInput) (*provid.AuthenticateOutput, error) {
	p.logger.Info("authenticating user")

	if err := p.resolveConfig(input.ConfigSet); err != nil {
		return nil, fmt.Errorf("resolving config: %w", err)
	}

	cfg := &aadConfig{}
	if err := config.Unmarshall(input.ConfigSet, cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into use aadconfig: %w", err)
	}

	if err := p.validateConfig(cfg); err != nil {
		return nil, err
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
		token, err = identityClient.GetOauth2TokenFromUsernamePassword(authCfg, "https://management.azure.com/")
		if err != nil {
			return nil, err //TODO: specific error
		}
	case identity.AccountTypeUnknown: //TODO: need to check other parts
		if userRealm.CloudInstanceName == "microsoftonline.com" && userRealm.CloudAudienceURN == "urn:federation:MicrosoftOnline" {
			token, err = identityClient.GetOauth2TokenFromUsernamePassword(authCfg, "https://management.azure.com/")
			if err != nil {
				return nil, err //TODO: specific error
			}
		} else {
			return nil, identity.ErrUnknownAccountType
		}
	default:
		return nil, identity.ErrUnknownAccountType
	}

	id, err := p.createIdentityFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("creating identity from oauth token: %w", err)
	}

	return &provid.AuthenticateOutput{
		Identity: id,
	}, nil
}

func (p *aadIdentityProvider) createIdentityFromToken(token *identity.OauthToken) (*oidc.Identity, error) {
	id := &oidc.Identity{
		AccessToken:    token.AccessToken,
		IDToken:        token.IDToken,
		RefreshToken:   token.RefreshToken,
		Resource:       token.Resource,
		Scope:          token.Scope,
		IDProviderName: ProviderName,
	}

	switch {
	case token.ExpiresOn != "":
		expiresSeconds, err := token.ExpiresOn.Int64()
		if err != nil {
			return nil, fmt.Errorf("getting ExpiresOn: %w", err)
		}
		id.Expires = time.Unix(expiresSeconds, 0)
	case token.ExpiresIn != "":
		expiresInSeconds := token.ExpiresIn.String()

		numSeconds, err := strconv.Atoi(expiresInSeconds)
		if err != nil {
			return nil, fmt.Errorf("converting ExpiresIn: %w", err)
		}
		id.Expires = time.Now().Add(time.Second * time.Duration(numSeconds)).UTC()
	default:
		return nil, ErrNoExpiryTime
	}

	return id, nil
}

func (p *aadIdentityProvider) validateConfig(cfg *aadConfig) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validating aad config: %w", err)
	}
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

	return cs, nil
}
