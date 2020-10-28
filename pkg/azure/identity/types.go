package identity

type Client interface {
	GetUserRealm(cfg *AuthenticationConfig) (*UserRealm, error)
	GetMex(federationMetadataURL string) (*MexDocument, error)
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
)
