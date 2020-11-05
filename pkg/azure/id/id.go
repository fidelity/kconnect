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
	ErrIDUnrecognizedFormat = errors.New("cluster id format is unrecognized")
)

type AzureIdentifier struct {
	SubscriptionID    string
	ResourceGroupName string
	Provider          string
	ResourceType      string
	ResourceName      string
}

// Parse will create a AzureIdentifier from a id string
func Parse(resourceID string) (*AzureIdentifier, error) {
	if strings.HasPrefix(resourceID, "/") {
		resourceID = resourceID[1:]
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) != 8 {
		return nil, ErrIDUnrecognizedFormat
	}

	subscription := parts[0]
	subscriptionID := parts[1]
	if subscription != "subscriptions" {
		return nil, fmt.Errorf("finding subscription: %w", ErrIDUnrecognizedFormat)
	}

	resourceGroup := parts[2]
	resourceGroupName := parts[3]
	if resourceGroup != "resourcegroups" {
		return nil, fmt.Errorf("finding resource group: %w", ErrIDUnrecognizedFormat)
	}

	provider := parts[4]
	providerName := parts[5]
	if provider != "providers" {
		return nil, fmt.Errorf("finding provider: %w", ErrIDUnrecognizedFormat)
	}

	azureIdentifier := &AzureIdentifier{
		SubscriptionID:    subscriptionID,
		ResourceGroupName: resourceGroupName,
		Provider:          providerName,
		ResourceType:      parts[6],
		ResourceName:      parts[7],
	}

	return azureIdentifier, nil
}
