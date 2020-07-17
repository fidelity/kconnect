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
	"errors"
	"fmt"

	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/spf13/pflag"
)

var (
	errNoRoleArnFlag = errors.New("no role-arn flag found in resolver")
)

// FlagsResolver is used to resolve the values for AWS related flags interactively.
type FlagsResolver struct {
	identity providers.Identity
}

// Resolve will resolve the values for the AWS specific flags that have no value. It will
// query AWS and interactively ask the user for selections.
func (r *FlagsResolver) Resolve(identity providers.Identity, flags *pflag.FlagSet) error {
	r.identity = identity

	if err := r.resolveRoleArn(flags); err != nil {
		return fmt.Errorf("resolving role-arn: %w", err)
	}

	return nil
}

func (r *FlagsResolver) resolveRoleArn(flags *pflag.FlagSet) error {
	flag := flags.Lookup("role-arn")
	if flag == nil {
		return errNoRoleArnFlag
	}
	if flag.Value.String() != "" {
		return nil
	}

	return nil
}
