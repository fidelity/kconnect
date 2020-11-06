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

func (p *aksClusterProvider) Discover(ctx *provider.Context, identity provider.Identity) (*provider.DiscoverOutput, error) {
	if err := p.setup(ctx.ConfigurationItems(), identity); err != nil {
		return nil, fmt.Errorf("setting up aks provider: %w", err)
	}

	p.logger.Info("discovering AKS clusters")

	clusters, err := p.listClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}

	discoverOutput := &provider.DiscoverOutput{
		ClusterProviderName:  "aks",
		IdentityProviderName: "aad", //TODO: get this from the identity
		Clusters:             make(map[string]*provider.Cluster),
	}
	for _, v := range clusters {
		discoverOutput.Clusters[v.ID] = v
	}

	return discoverOutput, nil
}

func (p *aksClusterProvider) listClusters(ctx *provider.Context) ([]*provider.Cluster, error) {
	p.logger.Debugw("listing clusters", "subscription", *p.config.SubscriptionID)
	client := containerservice.NewManagedClustersClient(*p.config.SubscriptionID)
	client.Authorizer = p.authorizer

	clusters := []*provider.Cluster{}
	list, err := client.List(ctx.Context)
	if err != nil {
		return nil, fmt.Errorf("querying for AKS clusters: %w", err)
	}

	for _, val := range list.Values() {
		clusterID, err := id.ToClusterID(*val.ID)
		if err != nil {
			return nil, fmt.Errorf("create cluster id: %w", err)
		}
		cluster := &provider.Cluster{
			Name: *val.Name,
			ID:   clusterID,
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}
