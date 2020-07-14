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

package identity

import (
	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/spf13/pflag"
)

// EmptyIdentityProvider is an identity provider that does nothing. Its used for testing
type EmptyIdentityProvider struct {
}

// EmptyIdentity is a empty empty returned from the EmptyIdentityProvider
type EmptyIdentity struct {
}

// Name returns the name of the plugin
func (p *EmptyIdentityProvider) Name() string {
	return "Empty Identity Provider"
}

// Flags will return the flags for this plugin
func (p *EmptyIdentityProvider) Flags() *pflag.FlagSet {
	return nil
}

// Authenticate will authenticate a user and returns their identity
func (p *EmptyIdentityProvider) Authenticate() (providers.Identity, error) {
	return &EmptyIdentity{}, nil
}
