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

package id_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v5"
	. "github.com/onsi/gomega"

	"github.com/fidelity/kconnect/pkg/azure/id"
)

func TestClusterResourceConvFuzz(t *testing.T) {
	g := NewWithT(t)

	for i := 0; i < 100; i++ {
		original := &id.ResourceIdentifier{
			Provider:          id.ContainerServiceProvider,
			ResourceType:      id.ManagedClustersResource,
			SubscriptionID:    gofakeit.UUID(),
			ResourceGroupName: gofakeit.Word(),
			ResourceName:      gofakeit.Word(),
		}

		clusterID, err := id.ToClusterID(original.String())
		g.Expect(err).ToNot(HaveOccurred())

		final, err := id.FromClusterID(clusterID)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(final).To(BeEquivalentTo(original))
	}
}
