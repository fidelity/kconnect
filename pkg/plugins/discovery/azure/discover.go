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

	azclient "github.com/fidelity/kconnect/pkg/azure/client"
	"github.com/fidelity/kconnect/pkg/azure/id"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

func (p *aksClusterProvider) Discover(ctx context.Context, input *discovery.DiscoverInput) (*discovery.DiscoverOutput, error) {
	if err := p.setup(input.ConfigSet, input.Identity); err != nil {
		return nil, fmt.Errorf("setting up aks provider: %w", err)
	}
	p.logger.Info("discovering AKS clusters")

	clusters, err := p.listClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}

	discoverOutput := &discovery.DiscoverOutput{
		DiscoveryProvider: ProviderName,
		IdentityProvider:  input.Identity.IdentityProviderName(),
		Clusters:          make(map[string]*discovery.Cluster),
	}
	for _, v := range clusters {
		discoverOutput.Clusters[v.ID] = v
	}

	return discoverOutput, nil
}

func (p *aksClusterProvider) listClusters(ctx context.Context) ([]*discovery.Cluster, error) {
	p.logger.Debugw("listing clusters", "subscription", *p.config.SubscriptionID)
	client := azclient.NewContainerClient(*p.config.SubscriptionID, p.authorizer)

	clusters := []*discovery.Cluster{}
	var list containerservice.ManagedClusterListResultPage
	var err error
	if p.config.ResourceGroup == nil || *p.config.ResourceGroup == "" {
		list, err = client.List(ctx)
	} else {
		list, err = client.ListByResourceGroup(ctx, *p.config.ResourceGroup)
	}
	if err != nil {
		return nil, fmt.Errorf("querying for AKS clusters: %w", err)
	}

	for _, val := range list.Values() {
		if p.config.ClusterName == "" || p.config.ClusterName == *val.Name {
			clusterID, err := id.ToClusterID(*val.ID)
			if err != nil {
				return nil, fmt.Errorf("create cluster id: %w", err)
			}

			cluster := &discovery.Cluster{
				Name: *val.Name,
				ID:   clusterID,
			}

			controlPlaneEndpoint := ""
			if val.Fqdn != nil {
				controlPlaneEndpoint = fmt.Sprintf("https://%s:443", *val.Fqdn)
			}
			if val.PrivateFQDN != nil {
				controlPlaneEndpoint = fmt.Sprintf("https://%s:443", *val.PrivateFQDN)
			}
			if controlPlaneEndpoint != "" {
				cluster.ControlPlaneEndpoint = &controlPlaneEndpoint
			}

			clusters = append(clusters, cluster)
		}
	}

	return clusters, nil
}
