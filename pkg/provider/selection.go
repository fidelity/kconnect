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

package provider

import (
	"fmt"
	"sort"

	"github.com/fidelity/kconnect/pkg/prompt"
)

// SelectItemFunc is a function that is used abstract the method for selecting
// an item from the a list of possible values. Providers shouldn't directly
// ask the user for input and instead should use this.
type SelectItemFunc func(prompt string, items map[string]string) (string, error)

func DefaultItemSelection(promptMessage string, items map[string]string) (string, error) {
	options := make([]string, 0, len(items))

	for key := range items {
		options = append(options, key)
	}

	if len(options) == 1 {
		return items[options[0]], nil
	}

	sort.Strings(options)

	selectedItem, err := prompt.Choose("item", promptMessage, true, prompt.OptionsFromMap(items))
	if err != nil {
		return "", fmt.Errorf("selecting item: %w", err)
	}

	return items[selectedItem], nil
}
