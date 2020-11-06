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

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-09-01/containerservice"

	"github.com/fidelity/kconnect/pkg/azure/id"
	"github.com/fidelity/kconnect/pkg/provider"
)

// Get will get the details of a AKS cluster.
func (p *aksClusterProvider) Get(ctx *provider.Context, clusterID string, identity provider.Identity) (*provider.Cluster, error) {
	if err := p.setup(ctx.ConfigurationItems(), identity); err != nil {
		return nil, fmt.Errorf("setting up aks provider: %w", err)
	}
	p.logger.Infow("getting AKS cluster", "id", clusterID)

	resourceID, err := id.FromClusterID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("getting resource id: %w", err)
	}

	client := containerservice.NewManagedClustersClient(resourceID.SubscriptionID)
	client.Authorizer = p.authorizer

	result, err := client.Get(ctx.Context, resourceID.ResourceGroupName, resourceID.ResourceName)
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}

	cluster := &provider.Cluster{
		Name: *result.Name,
		ID:   clusterID,
	}

	return cluster, nil
}
