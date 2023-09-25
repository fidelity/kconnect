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

package oidc

const Oidc = "oidc"

// Identity represents an oidc identity
type Identity struct {
	OidcServer        string
	OidcId            string
	OidcSecret        string
	UsePkce           string
	SkipOidcTlsVerify string
}

func (i *Identity) Type() string {
	return Oidc
}

func (i *Identity) Name() string {
	return Oidc
}

func (i *Identity) IsExpired() bool {
	return true
}

func (i *Identity) IdentityProviderName() string {
	return Oidc
}
