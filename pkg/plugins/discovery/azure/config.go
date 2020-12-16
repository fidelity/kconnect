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

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-09-01/containerservice"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	azclient "github.com/fidelity/kconnect/pkg/azure/client"
	"github.com/fidelity/kconnect/pkg/azure/id"
	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *aksClusterProvider) GetClusterConfig(ctx *provider.Context, cluster *provider.Cluster, namespace string) (*api.Config, string, error) {
	p.logger.Debug("getting cluster config")

	cfg, err := p.getKubeconfig(ctx.Context, cluster)
	if err != nil {
		return nil, "", fmt.Errorf("getting kubeconfig: %w", err)
	}

	if namespace != "" {
		p.logger.Debugw("setting kubernetes namespace", "namespace", namespace)
		cfg.Contexts[cfg.CurrentContext].Namespace = namespace
	}

	return cfg, cfg.CurrentContext, nil
}

func (p *aksClusterProvider) getKubeconfig(ctx context.Context, cluster *provider.Cluster) (*api.Config, error) {
	resourceID, err := id.FromClusterID(cluster.ID)
	if err != nil {
		return nil, fmt.Errorf("parsing cluster id: %w", err)
	}

	client := azclient.NewContainerClient(resourceID.SubscriptionID, p.authorizer)

	var credentialList containerservice.CredentialResults
	if p.config.Admin {
		credentialList, err = client.ListClusterAdminCredentials(ctx, resourceID.ResourceGroupName, resourceID.ResourceName)
	} else {
		credentialList, err = client.ListClusterUserCredentials(ctx, resourceID.ResourceGroupName, resourceID.ResourceName)
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
