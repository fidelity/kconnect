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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"

	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

func (p *eksClusterProvider) Discover(ctx context.Context, input *discovery.DiscoverInput) (*discovery.DiscoverOutput, error) {
	if err := p.setup(input.ConfigSet, input.Identity); err != nil {
		return nil, fmt.Errorf("setting up eks provider: %w", err)
	}

	p.logger.Info("discovering EKS clusters")

	clusters, err := p.listClusters()
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}

	discoverOutput := &discovery.DiscoverOutput{
		DiscoveryProvider: ProviderName,
		IdentityProvider:  "aws",
		Clusters:          make(map[string]*discovery.Cluster),
	}

	if len(clusters) == 0 {
		p.logger.Info("no EKS clusters discovered")
		return discoverOutput, nil
	}

	for _, clusterName := range clusters {
		clusterDetail, err := p.getClusterConfig(clusterName)
		if err != nil {
			return nil, fmt.Errorf("getting cluster config: %w", err)
		}

		discoverOutput.Clusters[clusterDetail.ID] = clusterDetail
	}

	return discoverOutput, nil
}

func (p *eksClusterProvider) listClusters() ([]string, error) {
	input := &eks.ListClustersInput{}
	clusters := []string{}

	paginator := eks.NewListClustersPaginator(p.eksClient, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("listing clusters: %w", err)
		}

		clusters = append(clusters, page.Clusters...)
	}

	return clusters, nil
}

func (p *eksClusterProvider) getClusterConfig(clusterName string) (*discovery.Cluster, error) {
	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	output, err := p.eksClient.DescribeCluster(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("describing cluster %s: %w", clusterName, err)
	}

	return &discovery.Cluster{
		ID:                       *output.Cluster.Arn,
		Name:                     *output.Cluster.Name,
		ControlPlaneEndpoint:     output.Cluster.Endpoint,
		CertificateAuthorityData: output.Cluster.CertificateAuthority.Data,
	}, nil
}
