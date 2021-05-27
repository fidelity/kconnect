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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/azure/wstrust"
	"github.com/fidelity/kconnect/pkg/defaults"
	khttp "github.com/fidelity/kconnect/pkg/http"
)

const (
	DefaultTokenDurationMins = 10
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

	return userRealm, nil
}

func (c *AzureADClient) GetMex(federationMetadataURL string) (*wstrust.MexDocument, error) {
	resp, err := c.httpClient.Get(federationMetadataURL, nil) //TODO: standard headers
	if err != nil {
		return nil, fmt.Errorf("getting mex endpoint: %w", err)
	}

	if resp.ResponseCode() != http.StatusOK {
		return nil, ErrInvalidResponseCode
	}

	doc, err := wstrust.CreateWsTrustMexDocument(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("creating wstrust doc: %w", err)
	}

	return doc, nil
}

func (c *AzureADClient) GetWsTrustResponse(cfg *AuthenticationConfig, cloudAudienceURN string, endpoint *wstrust.Endpoint) (*WSTrustResponse, error) {
	envelopeBody, err := c.createEnvelope(cfg, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error creating envelope: %w", err)
	}

	headers := defaults.Headers(defaults.WithNoCache(), defaults.WithAcceptJSON())
	headers["Content-Type"] = "application/soap+xml"
	headers["return-client-request-id"] = "true"

	switch endpoint.EndpointVersion {
	case wstrust.Trust2005:
		headers["soapaction"] = wstrust.Trust2005Spec
	case wstrust.Trust13:
		headers["soapaction"] = wstrust.Trust13Spec
	default:
		headers["soapaction"] = "http://docs.oasis-open.org/ws-sx/ws-trust/200512/RST/Issue"
	}

	resp, err := c.httpClient.Post(endpoint.URL, envelopeBody, headers)
	if err != nil {
		return nil, fmt.Errorf("posting envelope: %w", err)
	}
	if resp.ResponseCode() !=  http.StatusOK {
		return nil, ErrInvalidResponseCode
	}
	zap.S().Debug(resp.Body())

	wsTrustResp := &WSTrustResponse{}
	if err := xml.Unmarshal([]byte(resp.Body()), wsTrustResp); err != nil {
		return nil,  fmt.Errorf("error unmarshaling response: %w", err)
	}

	return wsTrustResp, nil
}

func (c *AzureADClient) GetOauth2TokenFromSamlAssertion(cfg *AuthenticationConfig, assertion string, resource string) (*OauthToken, error) {

	assertionEncoded := base64.StdEncoding.EncodeToString([]byte(assertion))
	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("resource", resource)
	data.Set("scope", "openid")
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:saml1_1-bearer")
	data.Set("assertion", assertionEncoded)

	url := fmt.Sprintf("%soauth2/token", cfg.Authority.AuthorityURI)

	headers := make(map[string]string) //TODO: add common headers
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	resp, err := c.httpClient.Post(url, data.Encode(), headers)
	if err != nil {
		return nil, err
	}

	token := &OauthToken{}
	if err := json.Unmarshal([]byte(resp.Body()), token); err != nil {
		return nil, fmt.Errorf("unmarshalling oauth token: %w", err)
	}

	return token, err
}

func (c *AzureADClient) GetOauth2TokenFromUsernamePassword(cfg *AuthenticationConfig, resource string) (*OauthToken, error) {

	params := map[string]string{
		"grant_type":  "password",
		"username":    cfg.Username,
		"password":    cfg.Password,
		"client_id":   cfg.ClientID,
		"resource":    resource,
		"client_info": "1",
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"

	body := c.endcodeQueryParams(params)

	resp, err := c.httpClient.Post(cfg.Endpoints.TokenEndpoint, body, headers)
	if err != nil {
		return nil, err
	}

	if resp.ResponseCode() != http.StatusOK {
		oidcResp := &OIDCErrorResponse{}
		if err := json.Unmarshal([]byte(resp.Body()), oidcResp); err != nil {
			return nil, fmt.Errorf("unmarshalling oidc error response: %w", err)
		}
		return nil, oidcResp
	}

	token := &OauthToken{}
	if err := json.Unmarshal([]byte(resp.Body()), token); err != nil {
		return nil, fmt.Errorf("unmarshalling oauth token: %w", err)
	}

	return token, nil
}

func (c *AzureADClient) createEnvelope(cfg *AuthenticationConfig, endpoint *wstrust.Endpoint) (string, error) {

	messageID := uuid.New()

	schemaLocation := "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"

	var soapAction string
	var requestTrustNamespace string
	var keyType string
	var requestType string

	switch endpoint.EndpointVersion {
	case wstrust.Trust2005:
		soapAction = wstrust.Trust2005Spec
		requestTrustNamespace = "http://schemas.xmlsoap.org/ws/2005/02/trust"
		keyType = "http://schemas.xmlsoap.org/ws/2005/05/identity/NoProofKey"
		requestType = "http://schemas.xmlsoap.org/ws/2005/02/trust/Issue"
	case wstrust.Trust13:
		soapAction = wstrust.Trust13Spec
		requestTrustNamespace = "http://docs.oasis-open.org/ws-sx/ws-trust/200512"
		keyType = "http://docs.oasis-open.org/ws-sx/ws-trust/200512/Bearer"
		requestType = "http://docs.oasis-open.org/ws-sx/ws-trust/200512/Issue"
	default:
		soapAction = "http://docs.oasis-open.org/ws-sx/ws-trust/200512/RST/Issue"
		requestTrustNamespace = "http://docs.oasis-open.org/ws-sx/ws-trust/200512"
		keyType = "http://docs.oasis-open.org/ws-sx/ws-trust/200512/Bearer"
		requestType = "http://docs.oasis-open.org/ws-sx/ws-trust/200512/Issue"
	}

	start := time.Now().UTC()
	end := start.Add(time.Minute * time.Duration(DefaultTokenDurationMins))

	startFormatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.%02dZ",
		start.Year(), start.Month(), start.Day(),
		start.Hour(), start.Minute(), start.Second(), start.Nanosecond())

	endFormatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.%02dZ",
		end.Year(), end.Month(), end.Day(),
		end.Hour(), end.Minute(), end.Second(), end.Nanosecond())

	envelopeParams := EnvelopeParams {
		SchemaLocation: schemaLocation,
		SoapAction: soapAction,
		MessageID: messageID.URN(),
		EndpointURL: formatXMLText(endpoint.URL),
		Created: startFormatted,
		Expires: endFormatted,
		Username: formatXMLText(cfg.Username),
		Password: formatXMLText(cfg.Password),
		RequestTrustNamespace: requestTrustNamespace,
		KeyType: keyType,
		RequestType: requestType,
	}

	requestTemplate, err := template.New("envelope").Parse(
	"<s:Envelope xmlns:s='http://www.w3.org/2003/05/soap-envelope' xmlns:wsa='http://www.w3.org/2005/08/addressing' xmlns:wsu='{{.SchemaLocation}}'>" +
		"<s:Header>" +
			"<wsa:Action s:mustUnderstand='1'>{{.SoapAction}}</wsa:Action>" +
			"<wsa:messageID>{{.MessageID}}</wsa:messageID>" + 
			"<wsa:ReplyTo>" + 
				"<wsa:Address>http://www.w3.org/2005/08/addressing/anonymous</wsa:Address>" +
			"</wsa:ReplyTo>" + 
			"<wsa:To s:mustUnderstand='1'>{{.EndpointURL}}</wsa:To>" +
			"<wsse:Security s:mustUnderstand='1' xmlns:wsse='http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd'>" +
				"<wsu:Timestamp wsu:Id='_0'>" +
					"<wsu:Created>{{.Created}}</wsu:Created>" +
					"<wsu:Expires>{{.Expires}}</wsu:Expires>" +
				"</wsu:Timestamp>" +
				"<wsse:UsernameToken wsu:Id='ADALUsernameToken'>" +
					"<wsse:Username>{{.Username}}</wsse:Username>" +
					"<wsse:Password>{{.Password}}</wsse:Password>" +
				"</wsse:UsernameToken>" +
			"</wsse:Security>" +
		"</s:Header>" +
		"<s:Body>" +
			"<wst:RequestSecurityToken xmlns:wst='{{.RequestTrustNamespace}}'>" +
				"<wsp:AppliesTo xmlns:wsp='http://schemas.xmlsoap.org/ws/2004/09/policy'>" + 
					"<wsa:EndpointReference>" +
						"<wsa:Address>urn:federation:MicrosoftOnline</wsa:Address>" +
					"</wsa:EndpointReference>" +
				"</wsp:AppliesTo>" +
				"<wst:KeyType>{{.KeyType}}</wst:KeyType>" +
				"<wst:RequestType>{{.RequestType}}</wst:RequestType>" +
			"</wst:RequestSecurityToken>" +
		"</s:Body>" +
	"</s:Envelope>")

	if err != nil {
		return "", fmt.Errorf("error creating envelope template: %w", err)
	}

	var buf bytes.Buffer
	err = requestTemplate.Execute(&buf, envelopeParams)
	if err != nil {
		return "", fmt.Errorf("error executing envelope template: %w", err)
	}
	return buf.String(), nil
}

func formatXMLText(s string) string {
	buf := new(strings.Builder)
	xml.EscapeText(buf, []byte(s))
	return buf.String()
}

func (c *AzureADClient) endcodeQueryParams(params map[string]string) string {
	var buffer bytes.Buffer

	keys := []string{}

	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		if i > 0 {
			buffer.WriteString("&")
		}
		buffer.WriteString(url.QueryEscape(key))
		buffer.WriteString("=")
		buffer.WriteString(url.QueryEscape(params[key]))
	}

	return buffer.String()
}
