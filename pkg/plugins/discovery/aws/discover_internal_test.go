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
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
)

func TestClusterFromDescription(t *testing.T) {
	const arn = "arn:aws:eks:us-east-1:123456789012:cluster/dev"

	testCases := []struct {
		name        string
		cluster     *ekstypes.Cluster
		expectReady bool
	}{
		{
			name: "active cluster",
			cluster: &ekstypes.Cluster{
				Arn:                  aws.String(arn),
				Name:                 aws.String("dev"),
				Endpoint:             aws.String("https://ABC.gr7.us-east-1.eks.amazonaws.com"),
				CertificateAuthority: &ekstypes.Certificate{Data: aws.String("Q0VSVA==")},
				Status:               ekstypes.ClusterStatusActive,
			},
			expectReady: true,
		},
		{
			name: "cluster being created has no endpoint or certificate authority",
			cluster: &ekstypes.Cluster{
				Arn:    aws.String(arn),
				Name:   aws.String("dev"),
				Status: ekstypes.ClusterStatusCreating,
			},
		},
		{
			name: "certificate authority without data",
			cluster: &ekstypes.Cluster{
				Arn:                  aws.String(arn),
				Name:                 aws.String("dev"),
				Endpoint:             aws.String("https://ABC.gr7.us-east-1.eks.amazonaws.com"),
				CertificateAuthority: &ekstypes.Certificate{},
			},
		},
		{
			name: "no endpoint",
			cluster: &ekstypes.Cluster{
				Arn:                  aws.String(arn),
				Name:                 aws.String("dev"),
				CertificateAuthority: &ekstypes.Certificate{Data: aws.String("Q0VSVA==")},
			},
		},
		{
			name:    "no arn or name",
			cluster: &ekstypes.Cluster{},
		},
		{
			name:    "no cluster",
			cluster: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cluster, err := clusterFromDescription("dev", tc.cluster)

			if !tc.expectReady {
				if !errors.Is(err, ErrClusterNotReady) {
					t.Fatalf("expected ErrClusterNotReady, got %v", err)
				}

				if cluster != nil {
					t.Fatalf("expected no cluster, got %v", cluster)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cluster.ID != arn || cluster.Name != "dev" {
				t.Fatalf("unexpected cluster: %+v", cluster)
			}

			if *cluster.ControlPlaneEndpoint != "https://ABC.gr7.us-east-1.eks.amazonaws.com" || *cluster.CertificateAuthorityData != "Q0VSVA==" {
				t.Fatalf("unexpected connection details: %+v", cluster)
			}
		})
	}
}
