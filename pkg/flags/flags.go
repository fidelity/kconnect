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

package flags

import (
	"errors"

	"github.com/spf13/pflag"
)

var (
	// ErrFlagMissing is an error when there is no flag with a given name
	ErrFlagMissing = errors.New("flag missing")
)

// ExistsWithValue returns true if a flag exists in a flagset and has a value
// and that value is non-empty
func ExistsWithValue(name string, flags *pflag.FlagSet) bool {
	flag := flags.Lookup(name)
	if flag == nil {
		return false
	}

	if flag.Value == nil {
		return false
	}

	if flag.Value.String() == "" {
		return false
	}

	return true
}
