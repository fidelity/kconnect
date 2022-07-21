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
	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	rshared "github.com/fidelity/kconnect/pkg/rancher"
)

func (p *rancherClusterProvider) Validate(cfg config.ConfigurationSet) error {
	errsValidation := &kerrors.ValidationFailed{}

	for _, item := range cfg.GetAll() {
		if item.Required && !cfg.ExistsWithValue(item.Name) {
			errsValidation.AddFailure(fmt.Sprintf("%s is required", item.Name))
		}
	}

	if len(errsValidation.Failures()) > 0 {
		return errsValidation
	}

	return nil
}

// Resolve will resolve the values for the Rancher specific flags that have no value. It will
// query Rancher and interactively ask the user for selections.
func (p *rancherClusterProvider) Resolve(cfg config.ConfigurationSet, identity identity.Identity) error {
	if err := p.setup(cfg, identity); err != nil {
		return fmt.Errorf("setting up rancher provider: %w", err)
	}
	p.logger.Debug("resolving Rancher configuration items")

	if err := rshared.ResolveCommon(cfg); err != nil {
		return fmt.Errorf("resolving common Rancher config: %w", err)
	}

	return nil
}
