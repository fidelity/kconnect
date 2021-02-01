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

var (
	// ErrIDUnrecognizedFormat is an error if the format of the resource id is unrecognozed
	ErrIDUnrecognizedFormat = errors.New("resource id format is unrecognized")
)

// ResourceIdentifier represents the uniqie identifier for a resource
type ResourceIdentifier struct {
	SubscriptionID    string
	ResourceGroupName string
	Provider          string
	ResourceType      string
	ResourceName      string
}

// String retruns the string representation of the resource identifier
func (r *ResourceIdentifier) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourcegroups/%s/providers/%s/%s/%s",
		r.SubscriptionID,
		r.ResourceGroupName,
		r.Provider,
		r.ResourceType,
		r.ResourceName,
	)
}

// Parse will create a AzureIdentifier from a id string
func Parse(resourceID string) (*ResourceIdentifier, error) {
	resourceID = strings.TrimPrefix(resourceID, "/")

	parts := strings.Split(resourceID, "/")
	if len(parts) != resourceIDPartsNum {
		return nil, ErrIDUnrecognizedFormat
	}

	if parts[0] != "subscriptions" {
		return nil, fmt.Errorf("finding subscription: %w", ErrIDUnrecognizedFormat)
	}
	subscriptionID := parts[1]

	if parts[2] != "resourcegroups" {
		return nil, fmt.Errorf("finding resource group: %w", ErrIDUnrecognizedFormat)
	}
	resourceGroupName := parts[3]

	if parts[4] != "providers" {
		return nil, fmt.Errorf("finding provider: %w", ErrIDUnrecognizedFormat)
	}
	providerName := parts[5]

	azureIdentifier := &ResourceIdentifier{
		SubscriptionID:    subscriptionID,
		ResourceGroupName: resourceGroupName,
		Provider:          providerName,
		ResourceType:      parts[6],
		ResourceName:      parts[7],
	}

	return azureIdentifier, nil
}
