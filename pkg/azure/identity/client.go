package identity

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	khttp "github.com/fidelity/kconnect/pkg/http"
	"go.uber.org/zap"
)

type AzureADClient struct {
	httpClient khttp.Client
}

func NewClient(httpClient khttp.Client) Client {
	return &AzureADClient{
		httpClient: httpClient,
	}
}

func (c *AzureADClient) GetUserRealm(cfg *AuthenticationConfig) (*UserRealm, error) {
	url := fmt.Sprintf("https://%s/common/UserRealm/%s?api-version=1.0", cfg.Authority.Host, url.PathEscape(cfg.Username))

	resp, err := c.httpClient.Get(url, nil) //TODO: we should have standard headers?
	if err != nil {
		return nil, fmt.Errorf("getting user realm: %w", err)
	}

	if resp.ResponseCode() != http.StatusOK {
		return nil, ErrInvalidResponseCode
	}

	userRealm := &UserRealm{}

	if err := json.Unmarshal([]byte(resp.Body()), userRealm); err != nil {
		return nil, fmt.Errorf("unmarshalling user realm: %w", err)
	}

	//TODO: add validation???

	return userRealm, nil
}

func (c *AzureADClient) GetMex(federationMetadataURL string) (*MexDocument, error) {
	resp, err := c.httpClient.Get(federationMetadataURL, nil) //TODO: standard headers
	if err != nil {
		return nil, fmt.Errorf("getting mex endpoint: %w", err)
	}

	if resp.ResponseCode() != http.StatusOK {
		return nil, ErrInvalidResponseCode
	}
	fmt.Printf(resp.Body())

	mexDocument := &MexDocument{}
	if err := xml.Unmarshal([]byte(resp.Body()), mexDocument); err != nil {
		return nil, fmt.Errorf("unmarshalling mex document: %w", err)
	}

	for _, policy := range mexDocument.Policy {
		if policy.ExactlyOne.All.SignedSupportingTokens.Policy.UsernameToken.Policy.WssUsernameToken10 != "" {
			zap.S().Info("found username")
		}
	}

	return mexDocument, nil

}
