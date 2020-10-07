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

package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	expectedNameParts = 2
)

// Get will get the details of a EKS cluster. The clusterID maps to a ARN
func (p *eksClusterProvider) Get(ctx *provider.Context, clusterID string, identity provider.Identity) (*provider.Cluster, error) {
	p.logger.Infow("getting EKS cluster", "id", clusterID)

	if err := p.setup(ctx, identity); err != nil {
		return nil, fmt.Errorf("setting up eks provider: %w", err)
	}

	clusterName, err := p.getClusterName(clusterID)
	if err != nil {
		return nil, fmt.Errorf("getting cluster name for cluster id %s: %w", clusterID, err)
	}

	return p.getClusterConfig(clusterName)
}

func (p *eksClusterProvider) getClusterName(clusterID string) (string, error) {
	clusterARN, err := arn.Parse(clusterID)
	if err != nil {
		return "", fmt.Errorf("parsing cluster id as ARN: %w", err)
	}

	parts := strings.Split(clusterARN.Resource, "/")
	if len(parts) != expectedNameParts {
		return "", ErrUnexpectedClusterFormat
	}

	return parts[1], nil
}
