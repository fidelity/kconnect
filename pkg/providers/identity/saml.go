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

type SAMLIdentityProvider struct {
	idpEndpoint *string
	username    *string
	password    *string

	flags *pflag.FlagSet
}

type SAMLIdentity struct {
}

// Name returns the name of the plugin
func (p *SAMLIdentityProvider) Name() string {
	return "saml"
}

// Flags will return the flags for this plugin
func (p *SAMLIdentityProvider) Flags() *pflag.FlagSet {
	if p.flags == nil {
		p.flags = &pflag.FlagSet{}
		p.idpEndpoint = p.flags.String("idp-endpoint", "", "identity provider endpoint provided by your IT team")

		p.username = p.flags.String("username", "", "the username used for authentication")
		p.password = p.flags.String("password", "", "the password to use for authentication")
	}

	return p.flags
}

// Authenticate will authenticate a user and returns their identity
func (p *SAMLIdentityProvider) Authenticate() (providers.Identity, error) {
	return &SAMLIdentity{}, nil
}

// Usage returns a description for use in the help/usage
func (p *SAMLIdentityProvider) Usage() string {
	return "SAML Idp authentication"
}
