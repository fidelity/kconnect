package identity

import (
	"encoding/xml"

	"github.com/fidelity/kconnect/pkg/azure/wstrust"
)

type Client interface {
	GetUserRealm(cfg *AuthenticationConfig) (*UserRealm, error)
	GetMex(federationMetadataURL string) (*wstrust.MexDocument, error)
	GetWsTrustResponse(cfg *AuthenticationConfig, cloudAudienceURN string, endpoint *wstrust.Endpoint) (*WSTrustResponse, error)
	GetOauth2TokenFromSamlAssertion(cfg *AuthenticationConfig, assertion string, resource string) (*OauthToken, error)
	GetOauth2TokenFromUsernamePassword(cfg *AuthenticationConfig) (*OauthToken, error)
}

type AuthorityConfig struct {
	Tenant       string
	Host         AADHost
	AuthorityURI string
}

type AuthenticationConfig struct {
	Authority *AuthorityConfig
	ClientID  string
	Username  string
	Password  string
	Scopes    []string
	Endpoints *Endpoints
}

type Endpoints struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
}

//UserRealm is used to represent the details of a user
type UserRealm struct {
	AccountType           AccountType `json:"account_type"`
	DomainName            string      `json:"domain_name"`
	CloudInstanceName     string      `json:"cloud_instance_name"`
	CloudAudienceURN      string      `json:"cloud_audience_urn"`
	FederationProtocol    string      `json:"federation_protocol"`
	FederationMetadataURL string      `json:"federation_metadata_url"`
}

type AADHost string

var (
	AADHostWorldwide   = AADHost("login.microsoftonline.com")
	AADHostFallback    = AADHost("login.windows.net")
	AADHostChina       = AADHost("login.chinacloudapi.cn")
	AADHostGermany     = AADHost("login.microsoftonline.de")
	AADHostUSGov       = AADHost("login.microsoftonline.us")
	AADHostUSGovAPI    = AADHost("login.cloudgovapi.us")
	AADHostUSGovLegacy = AADHost("login-us.microsoftonline.com")
)

type AccountType string

var (
	AccountTypeFederated = AccountType("Federated")
	AccountTypeManaged   = AccountType("Managed")
	AccountTypeUnknown   = AccountType("Unknown")
)

type WSTrustResponse struct {
	XMLName xml.Name
	Body    WSTrustResponseBody
}

type WSTrustResponseBody struct {
	XMLName                                xml.Name                               `xml:"Body"`
	RequestSecurityTokenResponseCollection RequestSecurityTokenResponseCollection `xml:"RequestSecurityTokenResponseCollection"`
}

type RequestSecurityTokenResponseCollection struct {
	XMLName                      xml.Name                     `xml:"RequestSecurityTokenResponseCollection"`
	RequestSecurityTokenResponse RequestSecurityTokenResponse `xml:"RequestSecurityTokenResponse"`
}

type RequestSecurityTokenResponse struct {
	XMLName                xml.Name               `xml:"RequestSecurityTokenResponse"`
	RequestedSecurityToken RequestedSecurityToken `xml:"RequestedSecurityToken"`
}

type RequestedSecurityToken struct {
	Assertion string `xml:",innerxml"`
}

type OauthToken struct {
	Type         string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    string `json:"expires_in"` //seconds
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}
