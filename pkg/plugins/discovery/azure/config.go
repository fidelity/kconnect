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
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-03-01/containerservice"

	azclient "github.com/fidelity/kconnect/pkg/azure/client"
	"github.com/fidelity/kconnect/pkg/azure/id"
	azid "github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
)

const (
	AKSAADServerAppID = "6dae42f8-4368-4678-94ff-3960e28e3630"
)

func (p *aksClusterProvider) GetConfig(ctx context.Context, input *discovery.GetConfigInput) (*discovery.GetConfigOutput, error) {
	p.logger.Debug("getting cluster config")

	cfg, err := p.getKubeconfig(ctx, input.Cluster)
	if err != nil {
		return nil, fmt.Errorf("getting kubeconfig: %w", err)
	}
	if !p.config.Admin {
		if p.config.LoginType == LoginTypeToken {
			if err := p.addTokenToAuthProvider(cfg, input.Identity); err != nil {
				return nil, fmt.Errorf("adding oauth token to kubeconfig: %w", err)
			}
		} else {
			p.addKubelogin(cfg)
		}
		p.printLoginDetails()
	}

	if input.Namespace != nil && *input.Namespace != "" {
		p.logger.Debugw("setting kubernetes namespace", "namespace", *input.Namespace)
		cfg.Contexts[cfg.CurrentContext].Namespace = *input.Namespace
	}

	return &discovery.GetConfigOutput{
		KubeConfig:  cfg,
		ContextName: &cfg.CurrentContext,
	}, nil

}

func (p *aksClusterProvider) printLoginDetails() {
	if p.config.LoginType == LoginTypeResourceOwnerPassword {
		fmt.Fprintf(os.Stderr, "\033[33mSet the AAD_USER_PRINCIPAL_NAME and AAD_USER_PRINCIPAL_PASSWORD environment variables before running kubectl\033[0m\n")
	}
	if p.config.LoginType == LoginTypeServicePrincipal {
		fmt.Fprintf(os.Stderr, "\033[33mSet the AAD_SERVICE_PRINCIPAL_CLIENT_ID and AAD_SERVICE_PRINCIPAL_CLIENT_SECRET environment variables before running kubectl\033[0m\n")
	}

	if p.config.AzureEnvironment == EnvironmentStackCloud {
		fmt.Fprintf(os.Stderr, "\033[33mSet the Azure Stack URLs in a config file and set the AZURE_ENVIRONMENT_FILEPATH environment variable to the path of that file\033[0m\n")
	}
}

func (p *aksClusterProvider) addKubelogin(cfg *api.Config) {
	contextName := cfg.CurrentContext
	context := cfg.Contexts[contextName]
	userName := context.AuthInfo

	execConfig := &api.ExecConfig{
		APIVersion: "client.authentication.k8s.io/v1beta1",
		Command:    "kubelogin",
		Args: []string{
			"get-token",
			"--environment",
			mapAzureEnvironment(p.config.AzureEnvironment),
			"--server-id",
			AKSAADServerAppID,
			"--client-id",
			p.config.ClientID,
			"--tenant-id",
			p.config.TenantID,
			"--login",
			string(p.config.LoginType),
			"--server-fqdn-type",
			"public",
		},
	}

	cfg.AuthInfos = map[string]*api.AuthInfo{
		userName: {
			Exec: execConfig,
		},
	}
}

func (p *aksClusterProvider) addTokenToAuthProvider(cfg *api.Config, userID identity.Identity) error {
	id, ok := userID.(*azid.ActiveDirectoryIdentity)
	if !ok {
		return ErrTokenNeedsAD
	}

	contextName := cfg.CurrentContext
	context := cfg.Contexts[contextName]
	userName := context.AuthInfo
	authInfo := cfg.AuthInfos[userName]

	providerConfig := authInfo.AuthProvider.Config
	apiServerID := providerConfig["apiserver-id"]
	clientID := providerConfig["client-id"]

	updatedID := id.Clone(azid.WithClientID(clientID))

	token, err := updatedID.GetOAuthToken(apiServerID)
	if err != nil {
		return fmt.Errorf("getting oauth token for %s: %w", apiServerID, err)
	}

	providerConfig["access-token"] = token.AccessToken
	providerConfig["expires-in"] = token.ExpiresIn.String()
	providerConfig["expires-on"] = token.ExpiresOn.String()
	providerConfig["refresh-token"] = token.RefreshToken

	return nil
}

func (p *aksClusterProvider) getKubeconfig(ctx context.Context, cluster *discovery.Cluster) (*api.Config, error) {
	resourceID, err := id.FromClusterID(cluster.ID)
	if err != nil {
		return nil, fmt.Errorf("parsing cluster id: %w", err)
	}

	client := azclient.NewContainerClient(resourceID.SubscriptionID, p.authorizer)

	var credentialList containerservice.CredentialResults
	if p.config.Admin {
		credentialList, err = client.ListClusterAdminCredentials(ctx, resourceID.ResourceGroupName, resourceID.ResourceName, p.config.ServerFqdnType)
	} else {
		credentialList, err = client.ListClusterUserCredentials(ctx, resourceID.ResourceGroupName, resourceID.ResourceName, p.config.ServerFqdnType, "azure")
	}
	if err != nil {
		return nil, fmt.Errorf("getting user credentials: %w", err)
	}

	if credentialList.Kubeconfigs == nil || len(*credentialList.Kubeconfigs) < 1 {
		return nil, ErrNoKubeconfigs
	}

	config := *(*credentialList.Kubeconfigs)[0].Value
	kubeCfg, err := clientcmd.Load(config)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}

	return kubeCfg, err
}

func mapAzureEnvironment(env Environment) string {
	switch env {
	case EnvironmentPublicCloud:
		return "AzurePublicCloud"
	case EnvironmentChinaCloud:
		return "AzureChinaCloud"
	case EnvironmentUSGovCloud:
		return "AzureUSGovernmentCloud"
	case EnvironmentStackCloud:
		return "AzureStackCloud"
	default:
		return ""
	}
}
