/*
Copyright 2021 The kconnect Authors.

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

package resolvers

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"

	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/resolve"
)

func IdpProvider(item *config.Item, cs config.ConfigurationSet) error {
	//TODO: get this from saml2aws????
	options := []string{"Akamai", "AzureAD", "ADFS", "ADFS2", "GoogleApps", "Ping", "PingNTLM", "JumpCloud", "Okta", "OneLogin", "PSU", "KeyCloak", "F5APM", "Shibboleth", "ShibbolethECP", "NetIQ"}

	idpProvider := ""
	prompt := &survey.Select{
		Message: "Select your identity provider",
		Options: options,
	}
	if err := survey.AskOne(prompt, &idpProvider, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for idp-provider: %w", err)
	}

	if err := cs.SetValue(item.Name, idpProvider); err != nil {
		return fmt.Errorf("setting idp-provider config: %w", err)
	}

	return nil
}

func IdpEndpoint(item *config.Item, cs config.ConfigurationSet) error {
	// if cs.ValueIsList(item.Name) {
	// 	listName := cs.ValueString(item.Name)
	// 	return resolve.ChooseFromList(cs, item.Name, item.ResolutionPrompt, item.Required, listName)
	// } else {
	return resolve.Input(cs, item.Name, item.ResolutionPrompt, item.Required)
	//}
}

// Username will interactively resolve the username config item
func Username(item *config.Item, cs config.ConfigurationSet) error {
	username := ""
	prompt := &survey.Input{
		Message: "Enter your username",
	}
	if err := survey.AskOne(prompt, &username, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for username name: %w", err)
	}

	if err := cs.SetValue(defaults.UsernameConfigItem, username); err != nil {
		return fmt.Errorf("setting username config: %w", err)
	}

	return nil
}

// Password will interactively resolve the password config item
func Password(item *config.Item, cs config.ConfigurationSet) error {
	password := ""
	prompt := &survey.Password{
		Message: "Enter your password",
	}
	if err := survey.AskOne(prompt, &password, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for password name: %w", err)
	}

	if err := cs.SetValue(defaults.PasswordConfigItem, password); err != nil {
		return fmt.Errorf("setting password config: %w", err)
	}

	return nil
}

// Alias will ask if an alias should be used
func Alias(item *config.Item, cs config.ConfigurationSet) error {
	useAlias := false
	promptUse := &survey.Confirm{
		Message: "Do you want to set an alias?",
	}
	if err := survey.AskOne(promptUse, &useAlias); err != nil {
		return fmt.Errorf("asking if an alias should be set: %w", err)
	}

	if !useAlias {
		return nil
	}

	alias := ""
	promptAliasName := &survey.Input{
		Message: "Enter the alias name",
	}
	if err := survey.AskOne(promptAliasName, &alias); err != nil {
		return fmt.Errorf("asking for alias name: %w", err)
	}

	if err := cs.SetValue(defaults.AliasConfigItem, alias); err != nil {
		return fmt.Errorf("setting alias config: %w", err)
	}

	return nil
}
