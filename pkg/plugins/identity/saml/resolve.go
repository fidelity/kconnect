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

package saml

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *samlIdentityProvider) Resolve(cfg config.ConfigurationSet, identity provider.Identity) error {
	sp := p.serviceProvider

	if p.interactive {
		p.logger.Debug("resolving SAML provider flags")
		if err := sp.ResolveConfiguration(cfg); err != nil {
			return fmt.Errorf("resolving flags: %w", err)
		}
	} else {
		p.logger.Debug("skipping configuration resolution as runnning non-interactive")
	}

	return nil
}

func (p *samlIdentityProvider) Validate(cfg config.ConfigurationSet) error {
	//NOTE: we haven't included the service provider validation here
	// as its not initialized yet.
	return config.ValidateRequired(cfg)
}
