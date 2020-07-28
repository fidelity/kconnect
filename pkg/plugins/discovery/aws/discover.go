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

	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"

	"github.com/fidelity/kconnect/pkg/aws"
	idprov "github.com/fidelity/kconnect/pkg/plugins/identity/saml"
	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *eksClusterProvider) Discover(ctx *provider.Context, identity provider.Identity) (*provider.DiscoverOutput, error) {
	logger := ctx.Logger().WithField("provider", "eks")
	logger.Info("discovering EKS clusters AWS")

	awsID := identity.(*idprov.AWSIdentity)

	logger.Debugf("creating AWS session with region %s and profile %s", *p.region, awsID.ProfileName)
	session, err := aws.NewSession(*p.region, awsID.ProfileName)
	if err != nil {
		return nil, fmt.Errorf("getting aws session: %w", err)
	}
	eksClient := aws.NewEKSClient(session)

	clusters, err := p.listClusters(eksClient)
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}

	discoverOutput := &provider.DiscoverOutput{
		ClusterProviderName:  "eks",
		IdentityProviderName: "",
		Clusters:             make(map[string]*provider.Cluster),
	}

	if len(clusters) == 0 {
		logger.Info("no EKS clusters discovered")
		return discoverOutput, nil
	}

	for _, clusterName := range clusters {
		clusterDetail, err := p.getClusterConfig(*clusterName, eksClient)
		if err != nil {
			return nil, fmt.Errorf("getting cluster config: %w", err)
		}
		discoverOutput.Clusters[clusterDetail.Name] = clusterDetail

	}

	return discoverOutput, nil
}

func (p *eksClusterProvider) listClusters(client eksiface.EKSAPI) ([]*string, error) {
	input := &eks.ListClustersInput{}

	//TODO: handle cluster name flag

	clusters := []*string{}
	err := client.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		clusters = append(clusters, page.Clusters...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}

	return clusters, nil
}

func (p *eksClusterProvider) getClusterConfig(clusterName string, client eksiface.EKSAPI) (*provider.Cluster, error) {

	input := &eks.DescribeClusterInput{
		Name: awsgo.String(clusterName),
	}

	output, err := client.DescribeCluster(input)
	if err != nil {
		return nil, fmt.Errorf("describing cluster %s: %w", clusterName, err)
	}

	return &provider.Cluster{
		ID:                       *output.Cluster.Arn,
		Name:                     *output.Cluster.Name,
		ControlPlaneEndpoint:     output.Cluster.Endpoint,
		CertificateAuthorityData: output.Cluster.CertificateAuthority.Data,
	}, nil
}
