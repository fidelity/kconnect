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

package aws

import (
	"fmt"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *eksClusterProvider) Validate(cfg config.ConfigurationSet) error {
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

// Resolve will resolve the values for the AWS specific flags that have no value. It will
// query AWS and interactively ask the user for selections.
func (p *eksClusterProvider) Resolve(cfg config.ConfigurationSet, userID provider.Identity) error {
	if err := kaws.ResolvePartition(cfg); err != nil {
		return fmt.Errorf("resolving partition: %w", err)
	}
	if err := kaws.ResolveRegion(cfg); err != nil {
		return fmt.Errorf("resolving region: %w", err)
	}
	return nil
}
