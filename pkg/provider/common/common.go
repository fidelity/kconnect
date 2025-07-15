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

package common

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
)

// ClusterProviderConfig represents the base configuration for
// a cluster provider
type ClusterProviderConfig struct {
	// ClusterId is the id of a cluster. The id should be unique for
	// the cluster provider
	ClusterID *string `json:"cluster-id"`

	// Alias is a friendly name to give to a cluster
	Alias *string `json:"alias"`
}

// IdentityProviderConfig represents the base configuration for an
// identity provider.
type IdentityProviderConfig struct {
	Username    string `json:"username"     validate:"required"`
	Password    string `json:"password"     validate:"required"`
	IdpProtocol string `json:"idp-protocol" validate:"required"`
}

func AddCommonClusterConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("cluster-id", "", "Id of the cluster to use."); err != nil {
		return fmt.Errorf("adding cluster-id setting: %w", err)
	}

	if _, err := cs.String("alias", "", "Friendly name to give to give the connection"); err != nil {
		return fmt.Errorf("adding alias setting: %w", err)
	}

	if err := cs.SetShort("cluster-id", "c"); err != nil {
		return fmt.Errorf("setting shorthand for cluster-id setting: %w", err)
	}

	if err := cs.SetShort("alias", "a"); err != nil {
		return fmt.Errorf("setting shorthand for alias setting: %w", err)
	}

	if err := cs.SetSensitive("alias"); err != nil {
		return fmt.Errorf("setting alias as sensitive: %w", err)
	}

	cs.SetHistoryIgnore("alias") //nolint: errcheck

	return nil
}

// AddCommonIdentityConfig will add common identity related config
func AddCommonIdentityConfig(cs config.ConfigurationSet) error {
	newCS := IdentityConfig()

	if err := cs.AddSet(newCS); err != nil {
		return fmt.Errorf("adding common identity config items: %w", err)
	}

	return nil
}

// IdentityConfig creates a configset with the common identity config items
func IdentityConfig() config.ConfigurationSet {
	cs := config.NewConfigurationSet()
	cs.String("username", "", "The username used for authentication")                                      //nolint: errcheck
	cs.String("password", "", "The password to use for authentication")                                    //nolint: errcheck
	cs.String("idp-protocol", "", "The idp protocol to use (e.g. saml). Each protocol has its own flags.") //nolint: errcheck
	cs.SetSensitive("password")                                                                            //nolint: errcheck

	return cs
}
