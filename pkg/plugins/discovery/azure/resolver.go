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
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-11-01/subscriptions"

	"github.com/fidelity/kconnect/pkg/config"
	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/provider"
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
func (p *aksClusterProvider) Resolve(config config.ConfigurationSet, identity provider.Identity) error {
	if err := p.setup(config, identity); err != nil {
		return fmt.Errorf("setting up aks provider: %w", err)
	}
	p.logger.Debug("resolving Azure configuration items")

	if err := p.resolveSubscription("subscription-id", config); err != nil {
		return fmt.Errorf("resolving subscription-id: %w", err)
	}

	return nil
}

func (p *aksClusterProvider) resolveSubscription(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	client := subscriptions.NewClient()
	client.Authorizer = p.authorizer

	res, err := client.List(context.TODO())
	if err != nil {
		return fmt.Errorf("getting subscription list: %w", err)
	}

	subs := make(map[string]string)
	options := []string{}
	for _, sub := range res.Values() {
		subs[*sub.DisplayName] = *sub.SubscriptionID
		options = append(options, *sub.DisplayName)
	}
	sort.Strings(options)

	if len(options) == 0 {
		return ErrNoSubscriptions
	}

	selectedSubscription := ""
	if len(options) == 1 {
		p.logger.Debug("only 1 subscription, automatically selected")
		selectedSubscription = options[0]
	} else {

		prompt := &survey.Select{
			Message: "Select an Azure subscription",
			Options: options,
		}
		if err := survey.AskOne(prompt, &selectedSubscription, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("asking for subscription: %w", err)
		}
	}

	subscriptionID := subs[selectedSubscription]

	if err := cfg.SetValue(name, subscriptionID); err != nil {
		return fmt.Errorf("setting subscription-id config: %w", err)
	}
	p.logger.Debugw("selected subscription", "name", selectedSubscription, "id", subscriptionID)

	return nil
}
