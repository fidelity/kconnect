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

package oidc

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider/identity"
)

func (p *oidcClusterProvider) Validate(cfg config.ConfigurationSet) error {
	return nil
}

// Resolve will resolve the values for the OIDC specific flags that have no value. It will
// read config file and interactively ask the user for selections.
func (p *oidcClusterProvider) Resolve(config config.ConfigurationSet, userID identity.Identity) error {
	if err := p.setup(config, userID); err != nil {
		return fmt.Errorf("setting up oidc provider: %w", err)
	}
	return nil
}
