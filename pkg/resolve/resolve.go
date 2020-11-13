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

package resolve

import (
	"fmt"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
)

// Username will interactively resolve the username config item
func Username(cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(defaults.UsernameConfigItem) {
		return nil
	}

	username := ""
	prompt := &survey.Input{
		Message: "Enter your username",
	}
	if err := survey.AskOne(prompt, &username, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for username name: %w", err)
	}

	if err := cfg.SetValue(defaults.UsernameConfigItem, username); err != nil {
		return fmt.Errorf("setting username config: %w", err)
	}

	return nil
}

// Password will interactively resolve the password config item
func Password(cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(defaults.PasswordConfigItem) {
		return nil
	}

	password := ""
	prompt := &survey.Password{
		Message: "Enter your password",
	}
	if err := survey.AskOne(prompt, &password, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for password name: %w", err)
	}

	if err := cfg.SetValue(defaults.PasswordConfigItem, password); err != nil {
		return fmt.Errorf("setting password config: %w", err)
	}

	return nil
}

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
