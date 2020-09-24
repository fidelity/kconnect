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

package config

import (
	"fmt"
	"strconv"
)

// ApplyToConfigSet will apply the saved app configuration to the supplied config set.
func ApplyToConfigSet(cs ConfigurationSet) error {
	return applyConfiguration(cs, "")
}

// ApplyToConfigSetWithProvider will apply the saved app configuration to the supplied config set and
// will take into consideration provider specific overrides
func ApplyToConfigSetWithProvider(cs ConfigurationSet, provider string) error {
	return applyConfiguration(cs, provider)
}

func applyConfiguration(cs ConfigurationSet, provider string) error {
	appConfig, err := NewAppConfiguration()
	if err != nil {
		return fmt.Errorf("creating app config store: %w", err)
	}

	cfg, err := appConfig.Get()
	if err != nil {
		return fmt.Errorf("getting app config: %w", err)
	}

	for _, item := range cs.GetAll() {
		if item.HasValue() {
			continue
		}

		// apply provider specific value first
		providerValues, hasProvider := cfg.Spec.Providers[provider]
		if hasProvider {
			providerVal, hasProviderVal := providerValues[item.Name]
			if hasProviderVal {
				if err := setItemValue(item, providerVal); err != nil {
					return fmt.Errorf("setting item value for %s from provider config: %w", item.Name, err)
				}
				continue
			}
		}

		// apply global value if we have one
		globalVal, hasGlobalVal := cfg.Spec.Global[item.Name]
		if hasGlobalVal {
			if err := setItemValue(item, globalVal); err != nil {
				return fmt.Errorf("setting item value for %s from global config: %w", item.Name, err)
			}
			continue
		}
	}

	return nil
}

func setItemValue(item *Item, value string) error {
	switch item.Type {
	case ItemTypeString:
		item.Value = value
		return nil
	case ItemTypeInt:
		intVal, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fmt.Errorf("parsing config as int: %w", err)
		}
		item.Value = intVal
		return nil
	case ItemTypeBool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parsing config as bool: %w", err)
		}
		item.Value = boolVal
		return nil
	default:
		return ErrUnknownItemType
	}
}
