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

package oidc

import (
	"context"

	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

func (p *oidcClusterProvider) GetCluster(ctx context.Context, input *discovery.GetClusterInput) (*discovery.GetClusterOutput, error) {
	cluster, _ := p.getCluster(ctx)
	return &discovery.GetClusterOutput{
		Cluster: cluster,
	}, nil
}
