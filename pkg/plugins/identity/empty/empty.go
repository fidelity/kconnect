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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/fidelity/kconnect/pkg/provider"
)

func init() {
	if err := provider.RegisterIdentityProviderPlugin("empty", newEmptyProvider()); err != nil {
		// TODO: handle fatal error
		log.Fatalf("Failed to register Empty identity provider plugin: %v", err)
	}
}

func newEmptyProvider() *emptyIdentityProvider {
	return &emptyIdentityProvider{}
}

// EmptyIdentityProvider is an identity provider that does nothing. Its used for testing
type emptyIdentityProvider struct {
}

// EmptyIdentity is a empty empty returned from the EmptyIdentityProvider
type EmptyIdentity struct {
}

// Name returns the name of the plugin
func (p *emptyIdentityProvider) Name() string {
	return "Empty Identity Provider"
}

// Flags will return the flags for this plugin
func (p *emptyIdentityProvider) Flags() *pflag.FlagSet {
	return nil
}

// Authenticate will authenticate a user and returns their identity
func (p *emptyIdentityProvider) Authenticate(ctx *provider.Context) (provider.Identity, error) {
	return &EmptyIdentity{}, nil
}

// Usage returns a description for use in the help/usage
func (p *emptyIdentityProvider) Usage() string {
	return "an identity provider that does nothing"
}
