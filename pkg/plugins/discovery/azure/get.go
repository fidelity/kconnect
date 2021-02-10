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

	azclient "github.com/fidelity/kconnect/pkg/azure/client"
	"github.com/fidelity/kconnect/pkg/azure/id"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

// Get will get the details of a AKS cluster.
func (p *aksClusterProvider) GetCluster(ctx context.Context, input *discovery.GetClusterInput) (*discovery.GetClusterOutput, error) {
	if err := p.setup(input.ConfigSet, input.Identity); err != nil {
		return nil, fmt.Errorf("setting up aks provider: %w", err)
	}
	p.logger.Infow("getting AKS cluster", "id", input.ClusterID)

	resourceID, err := id.FromClusterID(input.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("getting resource id: %w", err)
	}

	client := azclient.NewContainerClient(resourceID.SubscriptionID, p.authorizer)
	result, err := client.Get(ctx, resourceID.ResourceGroupName, resourceID.ResourceName)
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}

	cluster := &discovery.Cluster{
		Name: *result.Name,
		ID:   input.ClusterID,
	}

	return &discovery.GetClusterOutput{
		Cluster: cluster,
	}, nil
}
