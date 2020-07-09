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
	"github.com/urfave/cli/v2"
)

// IdentityProvider represents the interface used to implement an identity provider
// plugin. It provides authentication and authorization functionality.
// Initial plugins will be SAML, OAuth/OIDC, AzureAD
type IdentityProvider interface {
	// Name returns the name of the plugin
	Name() string

	// Flags will return the flags as part of a flagset for the plugin
	Flags() []cli.Flag

	// Authenticate will authenticate a user and return details of
	// their identity.
	Authenticate() (*Identity, error)
}

// NewIdentityProvider is a factory method for creating an identity provider
// based on name
func NewIdentityProvider(name string) *IdentityProvider {
	return nil
}

type Identity struct{}
