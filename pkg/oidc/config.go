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

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/prompt"
)

const (
	OidcServerConfigItem         = "oidc-server"
	OidcServerConfigDescription  = "oidc server url"
	OidcIdConfigItem             = "oidc-client-id"
	OidcIdConfigDescription      = "oidc client id"
	OidcSecretConfigItem         = "oidc-client-secret"
	OidcSecretConfigDescription  = "oidc client secret"
	UsePkceConfigItem            = "oidc-use-pkce"
	UsePkceConfigDescription     = "if use pkce"
	ClusterIdConfigItem          = "cluster-id"
	ClusterIdConfigDescription   = "cluster id"
	ClusterUrlConfigItem         = "cluster-url"
	ClusterUrlConfigDescription  = "cluster api server endpoint"
	ClusterAuthConfigItem        = "cluster-auth"
	ClusterAuthConfigDescription = "cluster auth data"
	ConfigUrlConfigItem          = "config-url"
	ConfigUrlConfigDescription   = "configuration endpoint"
	CaCertConfigItem             = "ca-cert"
	CaCertConfigDescription      = "ca cert for configuration url"
	SkipTlsVerifyConfigItem      = "skip-ssl"
	SkipTlsVerifyDescription     = "flag to skip ssl for calling config url"
	SkipOidcTlsVerifyConfigItem  = "skip-oidc-ssl"
	SkipOidcTlsVerifyDescription = "flag to skip ssl for calling oidc server"
)

// SharedConfig will return shared configuration items for OIDC based cluster and identity providers
func SharedConfig() config.ConfigurationSet {
	cs := config.NewConfigurationSet()
	cs.String(OidcServerConfigItem, "", OidcServerConfigDescription)         //nolint: errcheck
	cs.String(OidcIdConfigItem, "", OidcIdConfigDescription)                 //nolint: errcheck
	cs.String(OidcSecretConfigItem, "", OidcSecretConfigDescription)         //nolint: errcheck
	cs.String(UsePkceConfigItem, "", UsePkceConfigDescription)               //nolint: errcheck
	cs.String(ClusterUrlConfigItem, "", ClusterUrlConfigDescription)         //nolint: errcheck
	cs.String(ClusterAuthConfigItem, "", ClusterAuthConfigDescription)       //nolint: errcheck
	cs.String(ConfigUrlConfigItem, "", ConfigUrlConfigDescription)           //nolint: errcheck
	cs.String(CaCertConfigItem, "", CaCertConfigDescription)                 //nolint: errcheck
	cs.String(SkipTlsVerifyConfigItem, "", SkipTlsVerifyDescription)         //nolint: errcheck
	cs.String(SkipOidcTlsVerifyConfigItem, "", SkipOidcTlsVerifyDescription) //nolint: errcheck

	return cs
}

func ReadUserInput(key string, msg string) (string, error) {
	userInput, err := prompt.Input(key, msg, false)
	if userInput == "" || err != nil {
		return userInput, fmt.Errorf("error reading input for %s", key)
	}
	return userInput, nil
}
