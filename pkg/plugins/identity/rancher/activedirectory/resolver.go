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

package activedirectory

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/defaults"
	"github.com/fidelity/kconnect/pkg/prompt"
	rshared "github.com/fidelity/kconnect/pkg/rancher"
)

func (p *radIdentityProvider) resolveConfig(cfg config.ConfigurationSet) error {
	if !p.interactive {
		p.logger.Debug("skipping configuration resolution as runnning non-interactive")
		return nil
	}

	if err := prompt.InputAndSet(cfg, defaults.UsernameConfigItem, "Username:", true); err != nil {
		return fmt.Errorf("resolving %s: %w", defaults.UsernameConfigItem, err)
	}

	if err := prompt.InputSensitiveAndSet(cfg, defaults.PasswordConfigItem, "Password:", true); err != nil {
		return fmt.Errorf("resolving %s: %w", defaults.PasswordConfigItem, err)
	}

	if err := rshared.ResolveCommon(cfg); err != nil {
		return fmt.Errorf("resolving common Rancher config: %w", err)
	}

	return nil
}
