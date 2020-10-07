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

	awsclient "github.com/aws/aws-sdk-go/aws/client"

	"github.com/fidelity/kconnect/pkg/config"
	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/provider"
)

// NewConfigResolver creates a new config resolver for AWS
func NewConfigResolver(session awsclient.ConfigProvider) (provider.ConfigResolver, error) {
	if session == nil {
		return nil, ErrNoSession
	}

	return &awsConfigResolver{
		session: session,
	}, nil
}

// awsConfigResolver is used to resolve the config values for AWS interactively.
type awsConfigResolver struct {
	session awsclient.ConfigProvider
}

func (r *awsConfigResolver) Validate(cfg config.ConfigurationSet) error {
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
func (r *awsConfigResolver) Resolve(config config.ConfigurationSet) error {
	return nil
}
