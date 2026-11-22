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

package app_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

func discovered(clusters ...*discovery.Cluster) *discovery.DiscoverOutput {
	output := &discovery.DiscoverOutput{Clusters: map[string]*discovery.Cluster{}}
	for _, cluster := range clusters {
		output.Clusters[cluster.ID] = cluster
	}

	return output
}

func TestClusterOptionsWithUniqueNames(t *testing.T) {
	g := NewWithT(t)

	options := app.ClusterOptions(discovered(
		&discovery.Cluster{ID: "id-1", Name: "dev"},
		&discovery.Cluster{ID: "id-2", Name: "prod"},
	))

	g.Expect(options).To(Equal(map[string]string{"dev": "id-1", "prod": "id-2"}))
}

func TestClusterOptionsWithDuplicateNames(t *testing.T) {
	g := NewWithT(t)

	// For example, AKS clusters with the same name in different resource groups
	options := app.ClusterOptions(discovered(
		&discovery.Cluster{ID: "/subscriptions/s/resourceGroups/rg-a/cluster/aks", Name: "aks"},
		&discovery.Cluster{ID: "/subscriptions/s/resourceGroups/rg-b/cluster/aks", Name: "aks"},
		&discovery.Cluster{ID: "id-3", Name: "other"},
	))

	g.Expect(options).To(Equal(map[string]string{
		"aks (/subscriptions/s/resourceGroups/rg-a/cluster/aks)": "/subscriptions/s/resourceGroups/rg-a/cluster/aks",
		"aks (/subscriptions/s/resourceGroups/rg-b/cluster/aks)": "/subscriptions/s/resourceGroups/rg-b/cluster/aks",
		"other": "id-3",
	}))
}

func TestClusterOptionsEveryClusterCanBeSelected(t *testing.T) {
	g := NewWithT(t)

	output := discovered(
		&discovery.Cluster{ID: "a", Name: "same"},
		&discovery.Cluster{ID: "b", Name: "same"},
		&discovery.Cluster{ID: "c", Name: "same"},
		&discovery.Cluster{ID: "d", Name: "different"},
	)
	options := app.ClusterOptions(output)

	ids := []string{}
	for _, id := range options {
		ids = append(ids, id)
	}

	g.Expect(ids).To(ConsistOf("a", "b", "c", "d"))
}

func TestClusterOptionsWithNoClusters(t *testing.T) {
	g := NewWithT(t)

	g.Expect(app.ClusterOptions(discovered())).To(BeEmpty())
}

func TestDefaultSelectClusterWithOneCluster(t *testing.T) {
	g := NewWithT(t)

	// A single cluster is selected without asking, so no terminal is needed
	cluster, err := app.DefaultSelectCluster(discovered(&discovery.Cluster{ID: "id-1", Name: "dev"}))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cluster.ID).To(Equal("id-1"))
}
