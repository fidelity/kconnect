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

package rancher

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/prompt"
)

// ResolveCommon will interactively resolve the common configuration for rancher
func ResolveCommon(cfg config.ConfigurationSet) error {
	if err := prompt.InputAndSet(cfg, APIEndpointConfigName, "Enter the Rancher API endpoint", true); err != nil {
		return fmt.Errorf("resolving %s: %w", APIEndpointConfigName, err)
	}

	return nil
}
