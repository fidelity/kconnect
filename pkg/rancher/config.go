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
)

const (
	// APIEndpointConfigName is the name of the config item for the Rancher API endpoint
	APIEndpointConfigName = "api-endpoint"
)

// CommonConfig represents the common configuration for Rancher
type CommonConfig struct {
	// APIEndpoint is the URL for the rancher API endpoint
	APIEndpoint string `json:"api-endpoint" validate:"required"`
}

// AddCommonConfig adds the Rancher common configuration to a configuration set
func AddCommonConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String(APIEndpointConfigName, "", "The Rancher API endpoint"); err != nil {
		return fmt.Errorf("setting config item %s: %w", APIEndpointConfigName, err)
	}
	if err := cs.SetRequired(APIEndpointConfigName); err != nil {
		return fmt.Errorf("setting %s required: %w", APIEndpointConfigName, err)
	}

	return nil
}
