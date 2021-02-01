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
	"context"
	"fmt"

	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

// Get will get the details of a Rancher cluster.
func (p *rancherClusterProvider) GetCluster(ctx context.Context, input *discovery.GetClusterInput) (*discovery.GetClusterOutput, error) {
	if err := p.setup(input.ConfigSet, input.Identity); err != nil {
		return nil, fmt.Errorf("setting up rancher provider: %w", err)
	}
	p.logger.Infow("getting cluster via Rancher", "id", input.ClusterID)

	p.logger.Debugw("getting cluster details from Rancher api", "cluster", input.ClusterID)
	clusterDetail, err := p.getClusterDetails(input.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("getting cluster detail: %w", err)
	}

	cluster := &discovery.Cluster{
		Name: clusterDetail.Name,
		ID:   input.ClusterID,
	}

	return &discovery.GetClusterOutput{
		Cluster: cluster,
	}, nil
}
