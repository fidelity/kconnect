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

package client

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-09-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-11-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"

	"github.com/fidelity/kconnect/internal/version"
)

const (
	userAgentTemplate = "kconnect.fidelity.github.com/%s"
)

// NewContainerClient will create a new Azure container services client
func NewContainerClient(subscriptionID string, authorizer autorest.Authorizer) containerservice.ManagedClustersClient {
	azclient := containerservice.NewManagedClustersClient(subscriptionID)
	azclient.Authorizer = authorizer
	azclient.UserAgent = fmt.Sprintf(userAgentTemplate, version.Get().String())

	return azclient
}

// NewSubscriptionsClient will create a new Azure subscriptions client
func NewSubscriptionsClient(authorizer autorest.Authorizer) subscriptions.Client {
	subClient := subscriptions.NewClient()
	subClient.Authorizer = authorizer
	subClient.UserAgent = fmt.Sprintf(userAgentTemplate, version.Get().String())

	return subClient
}
