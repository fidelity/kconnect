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

package provider

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/utils"
)

func DefaultItemSelection(prompt string, items map[string]string) (string, error) {
	options := []string{}

	for key := range items {
		options = append(options, key)
	}
	if len(options) == 1 {
		return items[options[0]], nil
	}

	sort.Strings(options)
	selectedItem := ""
	selectPrompt := &survey.Select{
		Message: prompt,
		Options: options,
		Filter:  utils.SurveyFilter,
	}
	if err := survey.AskOne(selectPrompt, &selectedItem, survey.WithValidator(survey.Required)); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			zap.S().Info("Received interrupt, exiting..")
			os.Exit(0)
		}
		return "", fmt.Errorf("asking for role: %w", err)
	}

	return items[selectedItem], nil
}
