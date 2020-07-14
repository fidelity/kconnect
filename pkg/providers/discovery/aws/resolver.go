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
	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/spf13/pflag"
)

// AWSFlagsResolver is used to resolve the values for AWS related flags interactively.
type FlagsResolver struct {
}

// Resolve will resolve the values for the AWS specific flags that have no value. It will
// query AWS and interactively ask the user for selections.
func (r *FlagsResolver) Resolve(identity providers.Identity, flags *pflag.FlagSet) error {
	return nil
}
