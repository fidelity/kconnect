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

package sp

import (
	"github.com/spf13/pflag"
	"github.com/versent/saml2aws/pkg/cfg"

	"github.com/fidelity/kconnect/pkg/provider"
)

type ProviderConfig struct {
	provider.IdentityProviderConfig
	IdpEndpoint string `flag:"idp-endpoint" validate:"required"`
	IdpProvider string `flag:"idp-provider" validate:"required"`
}

type ServiceProvider interface {
	Validate(ctx *provider.Context, flags *pflag.FlagSet) error
	ResolveFlags(ctx *provider.Context, flags *pflag.FlagSet) error
	PopulateAccount(account *cfg.IDPAccount, flags *pflag.FlagSet) error
	ProcessAssertions(account *cfg.IDPAccount, samlAssertions string) (provider.Identity, error)
}
