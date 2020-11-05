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
	"errors"
	"fmt"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-09-01/containerservice"

	"github.com/fidelity/kconnect/pkg/azure/id"
	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *aksClusterProvider) GetClusterConfig(ctx *provider.Context, cluster *provider.Cluster, namespace string) (*api.Config, string, error) {
	p.logger.Debug("getting cluster config")

	resourceID, err := id.Parse(cluster.ID)
	if err != nil {
		return nil, "", fmt.Errorf("parsing cluster id: %w", err)
	}

	client := containerservice.NewManagedClustersClient(resourceID.SubscriptionID)
	client.Authorizer = p.authorizer

	result, err := client.ListClusterUserCredentials(ctx.Context, resourceID.ResourceGroupName, resourceID.ResourceName)
	if err != nil {
		return nil, "", fmt.Errorf("getting user credentials: %w", err)
	}

	if len(*result.Kubeconfigs) == 0 {
		return nil, "", errors.New("no kubeconfigs")
	}

	return nil, "", nil
}
