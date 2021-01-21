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

package prompt

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/utils"
)

// Input will ask the user to enter a value
func Input(name, message string, required bool) (string, error) {
	enteredValue := ""
	prompt := &survey.Input{
		Message: message,
	}
	opts := []survey.AskOpt{}

	if required {
		opts = append(opts, survey.WithValidator(survey.Required))
	}

	if err := survey.AskOne(prompt, &enteredValue, opts...); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			zap.S().Info("Received interrupt, exiting..")
			os.Exit(0)
		}
		return "", fmt.Errorf("asking for %s name: %w", name, err)
	}

	return enteredValue, nil
}

// InputAndSet will resolve and set a configuration item by asking the user to enter a value
func InputAndSet(cfg config.ConfigurationSet, name, message string, required bool) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	var enteredValue string
	var err error

	if cfg.ValueIsList(name) {
		enteredValue, err = Choose(name, message, required, OptionsFromConfigList(cfg.ValueString(name)))
	} else {
		enteredValue, err = Input(name, message, required)
	}
	if err != nil {
		return fmt.Errorf("asking for %s name: %w", name, err)
	}

	if err := cfg.SetValue(name, enteredValue); err != nil {
		return fmt.Errorf("setting %s config: %w", name, err)
	}
	zap.S().Debugw("resolved config item", "name", name, "value", enteredValue)

	return nil
}

// InputSensitiveAndSet will resolve and set a configuration item by asking the user to enter
// a value but it won't show the value eneterd
func InputSensitiveAndSet(cfg config.ConfigurationSet, name, message string, required bool) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	enteredValue := ""
	prompt := &survey.Password{
		Message: message,
	}
	opts := []survey.AskOpt{}

	if required {
		opts = append(opts, survey.WithValidator(survey.Required))
	}

	if err := survey.AskOne(prompt, &enteredValue, opts...); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			zap.S().Info("Received interrupt, exiting..")
			os.Exit(0)
		}
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

// OptionsFromMap will create a OptionsFunc from a map
func OptionsFromMap(values map[string]string) OptionsFunc {
	return func() (map[string]string, error) {
		return values, nil
	}
}

// OptionsFromStringSlice will create a OptionsFunc from a slice of strings
func OptionsFromStringSlice(values []string) OptionsFunc {
	return func() (map[string]string, error) {
		mapValues := map[string]string{}
		for _, val := range values {
			mapValues[val] = val
		}

		return mapValues, nil
	}
}

// OptionsFromConfigList will return an options func based on a list from the app config
func OptionsFromConfigList(listName string) OptionsFunc {
	if strings.HasPrefix(listName, config.ListPrefix) {
		listName = strings.ReplaceAll(listName, config.ListPrefix, "")
	}
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

// ChooseAndSet will resolve and set a configuration item by asking the user to select a value from a list
func ChooseAndSet(cfg config.ConfigurationSet, name, message string, required bool, optionsFn OptionsFunc) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	selected, err := Choose(name, message, required, optionsFn)
	if err != nil {
		return fmt.Errorf("choosing value for %s: %w", name, err)
	}

	if err := cfg.SetValue(name, selected); err != nil {
		return fmt.Errorf("setting %s config: %w", name, err)
	}
	zap.S().Debugw("resolved config item", "name", name, "value", selected)

	return nil
}

// Choose will resolve a configuration item by asking the user to select a value from a list
func Choose(name, message string, required bool, optionsFn OptionsFunc) (string, error) {
	options, err := optionsFn()
	if err != nil {
		return "", err
	}

	displayOptions := []string{}
	for k := range options {
		displayOptions = append(displayOptions, k)
	}
	sort.Strings(displayOptions)

	selectedOptionDisplay := ""

	if len(displayOptions) == 1 { //nolint: nestif
		// If there is only 1 item we auto select
		selectedOptionDisplay = displayOptions[0]
	} else {
		prompt := &survey.Select{
			Message: message,
			Options: displayOptions,
			Filter:  utils.SurveyFilter,
		}

		opts := []survey.AskOpt{}
		if required {
			opts = append(opts, survey.WithValidator(survey.Required))
		}

		if err := survey.AskOne(prompt, &selectedOptionDisplay, opts...); err != nil {
			if errors.Is(err, terminal.InterruptErr) {
				zap.S().Info("Received interrupt, exiting..")
				os.Exit(0)
			}
			return "", fmt.Errorf("asking for %s: %w", name, err)
		}
	}
	selectedValue := options[selectedOptionDisplay]

	return selectedValue, nil
}

// Input will ask the user to enter a value
func Confirm(name, message string, required bool) (bool, error) {
	confirmedValue := false
	prompt := &survey.Confirm{
		Message: message,
	}
	opts := []survey.AskOpt{}

	if required {
		opts = append(opts, survey.WithValidator(survey.Required))
	}

	if err := survey.AskOne(prompt, &confirmedValue, opts...); err != nil {
		return confirmedValue, fmt.Errorf("asking for %s name: %w", name, err)
	}

	return confirmedValue, nil
}
