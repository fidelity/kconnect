package identity

import (
	"encoding/json"
	"fmt"
	"net/http"

	khttp "github.com/fidelity/kconnect/pkg/http"
)

type EndpointsResolver interface {
	Resolve(cfg *AuthorityConfig) (*Endpoints, error)
}

func NewOIDCEndpointsResolver(httpClient khttp.Client) EndpointsResolver {
	return &oidcEndpointsResolver{
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
		return nil, fmt.Errorf("oidc error: %s", oidcResp.Error)
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
	}

	return endpoints, nil
}
