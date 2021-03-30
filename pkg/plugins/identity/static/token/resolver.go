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

package token

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *staticTokenIdentityProvider) Resolve(cfg config.ConfigurationSet, identity provider.Identity) error {
	if !p.interactive {
		p.logger.Debug("skipping configuration resolution as runnning non-interactive")
	}

	if err := prompt.InputAndSet(cfg, tokenConfigItem, "Enter authentication token", true); err != nil {
		return fmt.Errorf("resolving %s: %w", tokenConfigItem, err)
	}

	return nil
}

func (p *staticTokenIdentityProvider) Validate(cfg config.ConfigurationSet) error {
	return config.ValidateRequired(cfg)
}
