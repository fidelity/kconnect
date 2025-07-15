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

package id

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ContainerServiceProvider = "Microsoft.ContainerService"
	ManagedClustersResource  = "managedClusters"

	clusterIDPartsNum  = 3
	resourceIDPartsNum = 8
)

var (
	ErrNotContainerService   = errors.New("cluster is not for the container service")
	ErrNotManagedCluster     = errors.New("resource type is not a managed cluster")
	ErrUnrecognizedClusterID = errors.New("cluster id in unrecognized format")
)

// ToClusterID creates a cluster id based on an azure resource id
func ToClusterID(clusterResourceID string) (string, error) {
	resourceID, err := Parse(clusterResourceID)
	if err != nil {
		return "", fmt.Errorf("parsing cluster id: %w", err)
	}

	if resourceID.Provider != ContainerServiceProvider {
		return "", ErrNotContainerService
	}

	if resourceID.ResourceType != ManagedClustersResource {
		return "", ErrNotManagedCluster
	}

	generatedID := fmt.Sprintf("%s/%s/%s", resourceID.SubscriptionID, resourceID.ResourceGroupName, resourceID.ResourceName)

	return generatedID, nil
}

// FromClusterID will create a ResourceIdentifer from a cluster id
func FromClusterID(clusterID string) (*ResourceIdentifier, error) {
	parts := strings.Split(clusterID, "/")
	if len(parts) != clusterIDPartsNum {
		return nil, ErrUnrecognizedClusterID
	}

	return &ResourceIdentifier{
		Provider:          ContainerServiceProvider,
		ResourceType:      ManagedClustersResource,
		SubscriptionID:    parts[0],
		ResourceGroupName: parts[1],
		ResourceName:      parts[2],
	}, nil
}
