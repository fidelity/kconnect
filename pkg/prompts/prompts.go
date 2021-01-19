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

package prompts

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
)

// Input will resolve a configuration item by asking the user to enter a value
func Input(cfg config.ConfigurationSet, name, message string, required bool) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	enteredValue := ""
	prompt := &survey.Input{
		Message: message,
	}
	opts := []survey.AskOpt{}

	if required {
		opts = append(opts, survey.WithValidator(survey.Required))
	}

	if err := survey.AskOne(prompt, &enteredValue, opts...); err != nil {
		return fmt.Errorf("asking for %s name: %w", name, err)
	}

	if err := cfg.SetValue(name, enteredValue); err != nil {
		return fmt.Errorf("setting %s config: %w", name, err)
	}
	zap.S().Debugw("resolved config item", "name", name, "value", enteredValue)

	return nil
}

// OptionsFunc is a function that will return the list of options to select from in the form
// if a map that is displayname:value
type OptionsFunc func() (map[string]string, error)

// Choose will resolve a configuration item by asking the user to select a value from a list
func Choose(cfg config.ConfigurationSet, name, message string, required bool, optionsFn OptionsFunc) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	options, err := optionsFn()
	if err != nil {
		return err
	}

	displayOptions := []string{}
	for k := range options {
		displayOptions = append(displayOptions, k)
	}
	sort.Strings(displayOptions)

	selectedOptionDisplay := ""

	if len(displayOptions) == 1 {
		// If there is only 1 item we auto select
		selectedOptionDisplay = displayOptions[0]
	} else {
		prompt := &survey.Select{
			Message: message,
			Options: displayOptions,
		}

		opts := []survey.AskOpt{}
		if required {
			opts = append(opts, survey.WithValidator(survey.Required))
		}

		if err := survey.AskOne(prompt, &selectedOptionDisplay, opts...); err != nil {
			return fmt.Errorf("asking for %s: %w", name, err)
		}
	}

	selectedValue := options[selectedOptionDisplay]
	if err := cfg.SetValue(name, selectedValue); err != nil {
		return fmt.Errorf("setting %s config: %w", name, err)
	}
	zap.S().Debugw("resolved config item", "name", name, "value", selectedValue)

	return nil
}

// ChooseFromList will resolve a configuration item by asking the user to select a value from a names list
// that is defined in the application config.
func ChooseFromList(cfg config.ConfigurationSet, name, message string, required bool, listName string) error {
	if strings.HasPrefix(listName, config.ListPrefix) {
		listName = strings.Replace(listName, config.ListPrefix, "", -1)
	}
	return Choose(cfg, name, message, required, listOptions(listName))
}

// ListAsOptions will return an options func based on a list from the app config
func listOptions(listName string) OptionsFunc {
	return func() (map[string]string, error) {
		appConfig, err := config.NewAppConfiguration()
		if err != nil {
			return nil, fmt.Errorf("getting app configuration: %w", err)
		}
		appCfg, err := appConfig.Get()
		if err != nil {
			return nil, fmt.Errorf("reading app configuration: %w", err)
		}
		list, ok := appCfg.Spec.Lists[listName]
		if !ok {
			return nil, fmt.Errorf("getting list %s: %w", listName, err)
		}
		items := map[string]string{}
		for _, listItem := range list {
			items[listItem.Name] = listItem.Value
		}

		return items, nil
	}
}
