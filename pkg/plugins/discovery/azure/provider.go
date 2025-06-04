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

package azure

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/Azure/go-autorest/autorest"
	"github.com/go-playground/validator/v10"

	azid "github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
	"github.com/fidelity/kconnect/pkg/utils"
)

const (
	ProviderName = "aks"
	UsageExample = `  # Discover AKS clusters using Azure AD
  {{.CommandPath}} use aks --idp-protocol aad

  # Discover AKS clusters using file based credentials
  export AZURE_TENANT_ID="123455"
  export AZURE_CLIENT_ID="76849"
  export AZURE_CLIENT_SECRET="supersecret"
  {{.CommandPath}} use aks --idp-protocol az-env
`
)

func init() {
	if err := registry.RegisterDiscoveryPlugin(&registry.DiscoveryPluginRegistration{
		PluginRegistration: registry.PluginRegistration{
			Name:                   ProviderName,
			UsageExample:           UsageExample,
			ConfigurationItemsFunc: ConfigurationItems,
		},
		CreateFunc:                 New,
		SupportedIdentityProviders: []string{"aad", "az-env"},
	}); err != nil {
		zap.S().Fatalw("Failed to register AKS discovery plugin", "error", err)
	}
}

// New will create a new Azure discovery plugin
func New(input *provider.PluginCreationInput) (discovery.Provider, error) {
	if input.HTTPClient == nil {
		return nil, provider.ErrHTTPClientRequired
	}

	return &aksClusterProvider{
		logger:      input.Logger,
		interactive: input.IsInteractive,
		httpClient:  input.HTTPClient,
	}, nil
}

type aksClusterProviderConfig struct {
	common.ClusterProviderConfig
	SubscriptionID   *string     `json:"subscription-id"`
	SubscriptionName *string     `json:"subscription-name"`
	ResourceGroup    *string     `json:"resource-group"`
	Admin            bool        `json:"admin"`
	ClusterName      string      `json:"cluster-name"`
	TenantID         string      `json:"tenant-id"`
	ClientID         string      `json:"client-id"`
	LoginType        LoginType   `json:"login-type"`
	AzureEnvironment Environment `json:"azure-env"`
	ServerFqdnType   string      `json:"server-fqdn-type"`
}

type aksClusterProvider struct {
	config     *aksClusterProviderConfig
	authorizer autorest.Authorizer

	httpClient  khttp.Client
	interactive bool
	logger      *zap.SugaredLogger
}

func (p *aksClusterProvider) Name() string {
	return ProviderName
}

func (p *aksClusterProvider) setup(cs config.ConfigurationSet, userID identity.Identity) error {
	cfg := &aksClusterProviderConfig{}
	if err := config.Unmarshall(cs, cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into aksClusteProviderConfig: %w", err)
	}
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validating config struct: %w", err)
	}

	p.config = cfg

	// TODO: should we just return a AuthorizerIdentity from the aad provider?
	switch userID.(type) { //nolint:gocritic,gosimple
	case *azid.ActiveDirectoryIdentity:
		id := userID.(*azid.ActiveDirectoryIdentity) //nolint: gosimple
		p.logger.Debugw("creating bearer authorizer")
		bearerAuth, err := getBearerAuthFromIdentity(id, "https://management.azure.com/")
		if err != nil {
			return fmt.Errorf("getting bearer authorizer: %w", err)
		}
		p.authorizer = bearerAuth
	case *azid.AuthorizerIdentity:
		id := userID.(*azid.AuthorizerIdentity) //nolint: gosimple
		p.authorizer = id.Authorizer()
	default:
		return ErrUnsupportedIdentity
	}

	return nil
}

func (p *aksClusterProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *aksClusterProvider) CheckPreReqs() error {
	return utils.CheckKubeloginPrereq()
}

// ConfigurationItems returns the configuration items for this provider
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()

	cs.String(SubscriptionIDConfigItem, "", "The Azure subscription to use (specified by ID)")                                                                                                  //nolint: errcheck
	cs.String(SubscriptionNameConfigItem, "", "The Azure subscription to use (specified by name)")                                                                                              //nolint: errcheck
	cs.String(ResourceGroupConfigItem, "", "The Azure resource group to use")                                                                                                                   //nolint: errcheck
	cs.Bool(AdminConfigItem, false, "Generate admin user kubeconfig")                                                                                                                           //nolint: errcheck
	cs.String(ClusterNameConfigItem, "", "The name of the AKS cluster")                                                                                                                         //nolint: errcheck
	cs.String(LoginTypeConfigItem, string(LoginTypeDeviceCode), "The login method to use when connecting to the AKS cluster as a non-admin. Possible values: devicecode,spn,ropc,msi,azurecli") //nolint: errcheck
	cs.String(AzureEnvironmentConfigItem, string(EnvironmentPublicCloud), "The Azure environment the clusters are in. Possible values: public,china,usgov,stack")                               //nolint: errcheck
	cs.String(ServerFqdnTypeConfigItem, "public", "Connect to AKS cluster via Public/Private FQDN")

	cs.SetShort(ResourceGroupConfigItem, "r") //nolint: errcheck

	return cs, nil
}

func getBearerAuthFromIdentity(id *azid.ActiveDirectoryIdentity, resource string) (autorest.Authorizer, error) {
	token, err := id.GetOAuthToken(resource)
	if err != nil {
		return nil, fmt.Errorf("getting oauth token from identity: %w", err)
	}

	bearerAuth := azid.NewExplicitBearerAuthorizer(token.AccessToken)
	return bearerAuth, nil
}
