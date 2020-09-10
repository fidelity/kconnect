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
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
)

// ClusterProviderConfig represents the base configuration for
// a cluster provider
type ClusterProviderConfig struct {
	ClusterName *string `json:"cluster"`
}

// IdentityProviderConfig represents the base configuration for an
// identity provider.
type IdentityProviderConfig struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Force       bool   `json:"force"`
	IdpProtocol string `json:"idp-protocol" validate:"required"`
}

func AddCommonClusterConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("cluster", "", "Name of the cluster to use."); err != nil {
		return fmt.Errorf("adding cluster setting: %w", err)
	}
	if _, err := cs.Bool("non-interactive", false, "Run without interactive flag resolution. Defaults to false"); err != nil {
		return fmt.Errorf("adding non-interactive setting: %w", err)
	}
	if err := cs.SetShort("cluster", "c"); err != nil {
		return fmt.Errorf("setting shorthand for cluster setting: %w", err)
	}

	return nil
}

// AddCommonIdentityFlags will add common identity related flags to a command
func AddCommonIdentityConfig(cs config.ConfigurationSet) error {
	newCS := CommonIdentityConfig()

	if err := cs.AddSet(newCS); err != nil {
		return fmt.Errorf("adding common identity config items: %w", err)
	}

	return nil
}

// CommonIdentityConfig creates a configset with the common identity config items
func CommonIdentityConfig() config.ConfigurationSet {
	cs := config.NewConfigurationSet()
	cs.String("username", "", "the username used for authentication")                //nolint: errcheck
	cs.String("password", "", "the password to use for authentication")              //nolint: errcheck
	cs.Bool("force", false, "If true then we force authentication every invocation") //nolint: errcheck
	cs.String("idp-protocol", "", "the idp protocol to use (e.g. saml)")             //nolint: errcheck

	return cs
}
