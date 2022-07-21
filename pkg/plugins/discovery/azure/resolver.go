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

	azclient "github.com/fidelity/kconnect/pkg/azure/client"
	"github.com/fidelity/kconnect/pkg/config"
	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider/identity"
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

// Resolve will resolve the values for the Azure specific flags that have no value. It will
// query Azure and interactively ask the user for selections.
func (p *aksClusterProvider) Resolve(cfg config.ConfigurationSet, userID identity.Identity) error {
	if err := p.setup(cfg, userID); err != nil {
		return fmt.Errorf("setting up aks provider: %w", err)
	}
	p.logger.Debug("resolving Azure configuration items")

	if cfg.ExistsWithValue(SubscriptionIDConfigItem) && cfg.ExistsWithValue(SubscriptionNameConfigItem) {
		return ErrSubscriptionNameOrID
	}

	if err := p.resolveSubscripionName(cfg); err != nil {
		return fmt.Errorf("resolving subscription name: %w", err)
	}

	if err := prompt.ChooseAndSet(cfg, SubscriptionIDConfigItem, "Choose the Azure subscription", true, p.subscriptionOptions); err != nil {
		return fmt.Errorf("resolving %s: %w", SubscriptionIDConfigItem, err)
	}

	return nil
}

func (p *aksClusterProvider) resolveSubscripionName(cfg config.ConfigurationSet) error {
	if !cfg.ExistsWithValue(SubscriptionNameConfigItem) {
		return nil
	}

	cfgItem := cfg.Get(SubscriptionNameConfigItem)
	subscriptionName := cfgItem.Value.(string)

	options, err := p.subscriptionOptions()
	if err != nil {
		return fmt.Errorf("getting subscriptions: %w", err)
	}
	id, ok := options[subscriptionName]
	if !ok {
		return fmt.Errorf("looking up subscription %s: %w", subscriptionName, err)
	}

	if err := cfg.SetValue(SubscriptionIDConfigItem, id); err != nil {
		return fmt.Errorf("setting %s config item: %w", SubscriptionIDConfigItem, err)
	}

	// If we have subscription id then ignore subscription id for history
	idCfgItem := cfg.Get(SubscriptionIDConfigItem)
	idCfgItem.HistoryIgnore = true

	return nil
}

func (p *aksClusterProvider) subscriptionOptions() (map[string]string, error) {
	client := azclient.NewSubscriptionsClient(p.authorizer)

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
