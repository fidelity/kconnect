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

package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-11-01/subscriptions"

	"github.com/fidelity/kconnect/pkg/config"
	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/resolve"
)

func (p *aksClusterProvider) Validate(cfg config.ConfigurationSet) error {
	errsValidation := &kerrors.ValidationFailed{}

	for _, item := range cfg.GetAll() {
		if item.Required && !cfg.ExistsWithValue(item.Name) {
			errsValidation.AddFailure(fmt.Sprintf("%s is required", item.Name))
		}
	}

	if len(errsValidation.Failures()) > 0 {
		return errsValidation
	}

	return nil
}

// Resolve will resolve the values for the AWS specific flags that have no value. It will
// query AWS and interactively ask the user for selections.
func (p *aksClusterProvider) Resolve(cfg config.ConfigurationSet, identity provider.Identity) error {
	if err := p.setup(cfg, identity); err != nil {
		return fmt.Errorf("setting up aks provider: %w", err)
	}
	p.logger.Debug("resolving Azure configuration items")

	if err := resolve.Choose(cfg, SubscriptionIDConfigItem, "Choose the Azure AAD host", true, p.subscriptionOptions); err != nil {
		return fmt.Errorf("resolving %s: %w", SubscriptionIDConfigItem, err)
	}

	return nil
}

func (p *aksClusterProvider) subscriptionOptions() (map[string]string, error) {
	client := subscriptions.NewClient()
	client.Authorizer = p.authorizer

	res, err := client.List(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("getting subscription list: %w", err)
	}

	subs := make(map[string]string)
	for _, sub := range res.Values() {
		subs[*sub.DisplayName] = *sub.SubscriptionID
	}

	return subs, nil
}
