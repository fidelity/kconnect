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
	"fmt"

	"github.com/fidelity/kconnect/pkg/provider"
)

// Get will get the details of a Rancher cluster.
func (p *rancherClusterProvider) Get(ctx *provider.Context, clusterID string, identity provider.Identity) (*provider.Cluster, error) {
	if err := p.setup(ctx.ConfigurationItems(), identity); err != nil {
		return nil, fmt.Errorf("setting up rancher provider: %w", err)
	}
	p.logger.Infow("getting cluster via Rancher", "id", clusterID)

	p.logger.Debugw("getting cluster details from Rancher api", "cluster", clusterID)
	clusterDetail, err := p.getClusterDetails(clusterID)
	if err != nil {
		return nil, fmt.Errorf("getting cluster detail: %w", err)
	}

	cluster := &provider.Cluster{
		Name: clusterDetail.Name,
		ID:   clusterID,
	}

	return cluster, nil
}
