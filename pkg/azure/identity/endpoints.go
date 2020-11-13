package identity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	khttp "github.com/fidelity/kconnect/pkg/http"
)

const (
	OAuth2Template     = "oauth2/%s%s"
	VersionQueryString = "?api-version=1.0"
)

type EndpointsResolver interface {
	Resolve(cfg *AuthorityConfig) (*Endpoints, error)
}

func NewOIDCEndpointsResolver(httpClient khttp.Client) EndpointsResolver {
	return &oidcEndpointsResolver{
		httpClient: httpClient,
	}
}

func NewOAuthEndpointsResolver(httpClient khttp.Client) EndpointsResolver {
	return &oauthEndpointsResolver{
		httpClient: httpClient,
	}
}

type oidcEndpointsResolver struct {
	httpClient khttp.Client
}

type oidcResponse struct {
	Error            string `json:"error"`
	SubError         string `json:"suberror"`
	ErrorDescription string `json:"error_description"`
	ErrorCodes       []int  `json:"error_codes"`
	CorrelationID    string `json:"correlation_id"`
	Claims           string `json:"claims"`

	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	Issuer                string `json:"issuer"`
}

func (r *oidcEndpointsResolver) Resolve(cfg *AuthorityConfig) (*Endpoints, error) {
	url := fmt.Sprintf("%sv2.0/.well-known/openid-configuration", cfg.AuthorityURI)

	resp, err := r.httpClient.Get(url, nil) //TODO: common headers
	if err != nil {
		return nil, fmt.Errorf("getting oidc config: %w", err)
	}

	if resp.ResponseCode() != http.StatusOK {
		return nil, ErrInvalidResponseCode
	}

	oidcResp := &oidcResponse{}
	if err := json.Unmarshal([]byte(resp.Body()), oidcResp); err != nil {
		return nil, fmt.Errorf("unmarshalling oidc response: %w", err)
	}

	if oidcResp.Error != "" {
		return nil, fmt.Errorf("oidc response %s: %w", oidcResp.Error, ErrOIDCResponse)
	}

	if oidcResp.AuthorizationEndpoint == "" {
		return nil, ErrAuthorizationEndpointNotFound
	}
	if oidcResp.TokenEndpoint == "" {
		return nil, ErrTokenEndpointNotFound
	}
	if oidcResp.Issuer == "" {
		return nil, ErrIssuerNotFound
	}

	endpoints := &Endpoints{
		AuthorizationEndpoint: oidcResp.AuthorizationEndpoint,
		TokenEndpoint:         oidcResp.TokenEndpoint,
		DeviceCodeEndpoint:    strings.Replace(oidcResp.TokenEndpoint, "token", "devicecode", -1),
	}

	return endpoints, nil
}

type oauthEndpointsResolver struct {
	httpClient khttp.Client
}

func (r *oauthEndpointsResolver) Resolve(cfg *AuthorityConfig) (*Endpoints, error) {
	u, err := url.Parse(cfg.AuthorityURI)
	if err != nil {
		return nil, err
	}
	authorizeURL, err := u.Parse(fmt.Sprintf(OAuth2Template, "authorize", VersionQueryString))
	if err != nil {
		return nil, err
	}
	tokenURL, err := u.Parse(fmt.Sprintf(OAuth2Template, "token", VersionQueryString))
	if err != nil {
		return nil, err
	}
	deviceCodeURL, err := u.Parse(fmt.Sprintf(OAuth2Template, "devicecode", VersionQueryString))
	if err != nil {
		return nil, err
	}

	return &Endpoints{
		AuthorizationEndpoint: authorizeURL.String(),
		TokenEndpoint:         tokenURL.String(),
		DeviceCodeEndpoint:    deviceCodeURL.String(),
	}, nil
}
