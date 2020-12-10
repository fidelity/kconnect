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

package rancher

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fidelity/kconnect/internal/defaults"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/rancher"
)

func (p *rancherClusterProvider) Discover(ctx *provider.Context, identity provider.Identity) (*provider.DiscoverOutput, error) {
	if err := p.setup(ctx.ConfigurationItems(), identity); err != nil {
		return nil, fmt.Errorf("setting up rancher provider: %w", err)
	}

	p.logger.Info("discovering clusters via Rancher")

	id, ok := identity.(*provider.TokenIdentity)
	if !ok {
		return nil, provider.ErrNotTokenIdentity
	}

	clusters, err := p.listClusters()
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}

	discoverOutput := &provider.DiscoverOutput{
		ClusterProviderName:  ProviderName,
		IdentityProviderName: id.IdentityProviderName(),
		Clusters:             make(map[string]*provider.Cluster),
	}
	for _, v := range clusters {
		discoverOutput.Clusters[v.ID] = v
	}

	return discoverOutput, nil
}

func (p *rancherClusterProvider) listClusters() ([]*provider.Cluster, error) {
	p.logger.Debug("listing clusters using rancker api")

	clusters := []*provider.Cluster{}

	resolver, err := rancher.NewStaticEndpointsResolver(p.config.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("creating endpoint resolver: %w", err)
	}

	headers := defaults.Headers(defaults.WithJSON(), defaults.WithBearerAuth(p.token))
	httpClient := khttp.NewHTTPClient()

	resp, err := httpClient.Get(resolver.ClustersList(), headers)
	if err != nil {
		return nil, fmt.Errorf("getting clusters using api: %w", err)
	}

	if resp.ResponseCode() != http.StatusOK {
		return nil, ErrGettingClusters
	}

	listClustersResponse := &listClustersResponse{}
	if err := json.Unmarshal([]byte(resp.Body()), listClustersResponse); err != nil {
		return nil, fmt.Errorf("unmarshalling api response: %w", err)
	}

	for _, val := range listClustersResponse.Clusters {
		cluster := &provider.Cluster{
			Name: val.Name,
			ID:   val.ID,
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}
