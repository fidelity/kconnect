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

package identity

import (
	"github.com/Azure/go-autorest/autorest"

	khttp "github.com/fidelity/kconnect/pkg/http"
)

func NewAuthorizerIdentity(name, idProviderName string, authorizer autorest.Authorizer) *AuthorizerIdentity {
	return &AuthorizerIdentity{
		name:           name,
		authorizer:     authorizer,
		idProviderName: idProviderName,
	}
}

type AuthorizerIdentity struct {
	authorizer     autorest.Authorizer
	name           string
	idProviderName string
}

func (a *AuthorizerIdentity) Type() string {
	return "azure-authorizer"
}

func (a *AuthorizerIdentity) Name() string {
	return a.name
}

func (a *AuthorizerIdentity) IsExpired() bool {
	//TODO: how do we do this properly
	return false
}

func (a *AuthorizerIdentity) Authorizer() autorest.Authorizer {
	return a.authorizer
}

func (a *AuthorizerIdentity) IdentityProviderName() string {
	return a.idProviderName
}

func NewActiveDirectoryIdentity(authCfg *AuthenticationConfig, userRealm *UserRealm, idProviderName string, httpClient khttp.Client, interactiveLoginRequired bool) *ActiveDirectoryIdentity {
	return &ActiveDirectoryIdentity{
		//authorizer:     authorizer,
		idProviderName:           idProviderName,
		realm:                    userRealm,
		authCfg:                  authCfg,
		httpClient:               httpClient,
		interactiveLoginRequired: interactiveLoginRequired,
	}
}

type ActiveDirectoryIdentity struct {
	authCfg *AuthenticationConfig
	realm   *UserRealm

	idProviderName           string
	httpClient               khttp.Client
	interactiveLoginRequired bool
}

func (a *ActiveDirectoryIdentity) Type() string {
	return "azure-active-directory"
}

func (a *ActiveDirectoryIdentity) Name() string {
	return a.authCfg.Username
}

func (a *ActiveDirectoryIdentity) IsExpired() bool {
	//TODO: how do we do this properly
	return false
}

func (a *ActiveDirectoryIdentity) IdentityProviderName() string {
	return a.idProviderName
}

func (a *ActiveDirectoryIdentity) GetOAuthToken(resource string) (*OauthToken, error) {
	if len(resource) == 0 {
		return nil, ErrResourceRequired
	}

	var token *OauthToken
	var err error

	identityClient := NewClient(a.httpClient)

	switch a.realm.AccountType {
	case AccountTypeFederated:
		doc, err := identityClient.GetMex(a.realm.FederationMetadataURL)
		if err != nil {
			return nil, err //TODO: specific error
		}
		endpoint := doc.UsernamePasswordEndpoint
		wsTrustResponse, err := identityClient.GetWsTrustResponse(a.authCfg, a.realm.CloudAudienceURN, &endpoint)
		if err != nil {
			return nil, err //TODO: specific error
		}
		token, err = identityClient.GetOauth2TokenFromSamlAssertion(a.authCfg, wsTrustResponse.Body.RequestSecurityTokenResponseCollection.RequestSecurityTokenResponse.RequestedSecurityToken.Assertion, resource)
		if err != nil {
			return nil, err //TODO: specific error
		}

	case AccountTypeManaged:
		if a.interactiveLoginRequired {
			token, err = identityClient.GetOauth2TokenFromAzureAccessToken(a.authCfg, resource)
		} else {
			token, err = identityClient.GetOauth2TokenFromUsernamePassword(a.authCfg, resource)
		}
		if err != nil {
			return nil, err //TODO: specific error
		}
	case AccountTypeUnknown: //TODO: need to check other parts
		if a.realm.CloudInstanceName == "microsoftonline.com" && a.realm.CloudAudienceURN == "urn:federation:MicrosoftOnline" {
			if a.interactiveLoginRequired {
				token, err = identityClient.GetOauth2TokenFromAzureAccessToken(a.authCfg, resource)
			} else {
				token, err = identityClient.GetOauth2TokenFromUsernamePassword(a.authCfg, resource)
			}
			if err != nil {
				return nil, err //TODO: specific error
			}
		} else {
			return nil, ErrUnknownAccountType
		}
	default:
		return nil, ErrUnknownAccountType
	}

	return token, nil
}

func (a *ActiveDirectoryIdentity) Clone(opts ...CloneOption) *ActiveDirectoryIdentity {
	copyID := &ActiveDirectoryIdentity{
		authCfg: &AuthenticationConfig{
			Authority: &AuthorityConfig{
				Tenant:       a.authCfg.Authority.Tenant,
				Host:         a.authCfg.Authority.Host,
				AuthorityURI: a.authCfg.Authority.AuthorityURI,
			},
			ClientID: a.authCfg.ClientID,
			Username: a.authCfg.Username,
			Password: a.authCfg.Password,
			Scopes:   a.authCfg.Scopes,
		},
		realm: &UserRealm{
			AccountType:           a.realm.AccountType,
			DomainName:            a.realm.DomainName,
			CloudInstanceName:     a.realm.CloudInstanceName,
			CloudAudienceURN:      a.realm.CloudAudienceURN,
			FederationProtocol:    a.realm.FederationProtocol,
			FederationMetadataURL: a.realm.FederationMetadataURL,
		},
		idProviderName:           a.idProviderName,
		httpClient:               a.httpClient,
		interactiveLoginRequired: a.interactiveLoginRequired,
	}
	if a.authCfg.Endpoints != nil {
		copyID.authCfg.Endpoints = &Endpoints{
			AuthorizationEndpoint: a.authCfg.Endpoints.AuthorizationEndpoint,
			TokenEndpoint:         a.authCfg.Endpoints.TokenEndpoint,
			DeviceCodeEndpoint:    a.authCfg.Endpoints.DeviceCodeEndpoint,
		}
	}

	for _, opt := range opts {
		opt(copyID)
	}

	return copyID
}

// Clone represents an option to use when cloning a ActiveDirectoryIdentity
type CloneOption func(*ActiveDirectoryIdentity)

func WithClientID(clientID string) CloneOption {
	return func(a *ActiveDirectoryIdentity) {
		a.authCfg.ClientID = clientID
	}
}
