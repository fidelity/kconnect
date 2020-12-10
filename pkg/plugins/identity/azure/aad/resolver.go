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

package aad

import (
	"fmt"

	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/azure/identity"
	"github.com/fidelity/kconnect/pkg/plugins/discovery/azure"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/resolve"
)

func (p *aadIdentityProvider) resolveConfig(ctx *provider.Context) error {
	if !ctx.IsInteractive() {
		p.logger.Debug("skipping configuration resolution as runnning non-interactive")
	}

	cfg := ctx.ConfigurationItems()

	if err := resolve.Username(cfg); err != nil {
		return fmt.Errorf("resolving %s: %w", defaults.UsernameConfigItem, err)
	}
	if err := resolve.Password(cfg); err != nil {
		return fmt.Errorf("resolving %s: %w", defaults.PasswordConfigItem, err)
	}
	if err := resolve.Input(cfg, azure.TenantIDConfigItem, "Enter the Azure tenant ID", true); err != nil {
		return fmt.Errorf("resolving %s: %w", azure.TenantIDConfigItem, err)
	}
	if err := resolve.Input(cfg, azure.ClientIDConfigItem, "Enter the Azure client ID", true); err != nil {
		return fmt.Errorf("resolving %s: %w", azure.ClientIDConfigItem, err)
	}
	if err := resolve.Choose(cfg, azure.AADHostConfigItem, "Choose the Azure AAD host", true, aadHostOptions); err != nil {
		return fmt.Errorf("resolving %s: %w", azure.ClientIDConfigItem, err)
	}

	return nil
}

func aadHostOptions() (map[string]string, error) {
	return map[string]string{
		"Worldwide (recommended)": string(identity.AADHostWorldwide),
		"China":                   string(identity.AADHostChina),
		"Germany":                 string(identity.AADHostGermany),
		"US Gov":                  string(identity.AADHostUSGov),
		"US Gov (API)":            string(identity.AADHostUSGovAPI),
		"US Gov (Legacy)":         string(identity.AADHostUSGovLegacy),
		"Fallback":                string(identity.AADHostFallback),
	}, nil
}
