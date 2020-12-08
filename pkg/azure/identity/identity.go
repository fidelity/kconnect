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
	"github.com/Azure/go-autorest/autorest"
)

func NewAuthorizerIdentity(name string, authorizer autorest.Authorizer) *AuthorizerIdentity {
	return &AuthorizerIdentity{
		name:       name,
		authorizer: authorizer,
	}
}

type AuthorizerIdentity struct {
	authorizer autorest.Authorizer
	name       string
}

func (a *AuthorizerIdentity) Type() string {
	return "azure-authorizer"
}

func (a *AuthorizerIdentity) Name() string {
	return a.name
}

func (a *AuthorizerIdentity) IsExpired() bool {
	//TODO: how do we do this properly
	return false
}

func (a *AuthorizerIdentity) Authorizer() autorest.Authorizer {
	return a.authorizer
}
