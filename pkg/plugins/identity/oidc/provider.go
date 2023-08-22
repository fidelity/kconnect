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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/oidc"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"

	kconnectv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	khttp "github.com/fidelity/kconnect/pkg/http"
)

const (
	ProviderName = "oidc"
)

func init() {
	if err := registry.RegisterIdentityPlugin(&registry.IdentityPluginRegistration{
		PluginRegistration: registry.PluginRegistration{
			Name:                   ProviderName,
			UsageExample:           "",
			ConfigurationItemsFunc: ConfigurationItems,
		},
		CreateFunc: New,
	}); err != nil {
		zap.S().Fatalw("Failed to register OIDC identity plugin", "error", err)
	}
}

// New will create a new OIDC identity provider
func New(input *provider.PluginCreationInput) (identity.Provider, error) {
	return &oidcIdentityProvider{
		logger:      input.Logger,
		interactive: input.IsInteractice,
	}, nil
}

type oidcIdentityProvider struct {
	logger      *zap.SugaredLogger
	interactive bool
}

type providerConfig struct {
	OidcServer string `json:"oidc-server"`
	OidcId     string `json:"oidc-client-id"`
	OidcSecret string `json:"oidc-client-secret"`
	UsePkce    string `json:"oidc-use-pkce"`
}

func (p *oidcIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will generate authentication config.
func (p *oidcIdentityProvider) Authenticate(ctx context.Context, input *identity.AuthenticateInput) (*identity.AuthenticateOutput, error) {
	p.logger.Info("using oidc for authentication")

	p.getConfigFromUrl(input.ConfigSet)

	cfg := &providerConfig{}
	if err := config.Unmarshall(input.ConfigSet, cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into providerConfig: %w", err)
	}

	id := &oidc.Identity{
		OidcServer: cfg.OidcServer,
		OidcId:     cfg.OidcId,
		OidcSecret: cfg.OidcSecret,
		UsePkce:    cfg.UsePkce,
	}

	return &identity.AuthenticateOutput{
		Identity: id,
	}, nil
}

func (p *oidcIdentityProvider) getConfigFromUrl(configSet config.ConfigurationSet) {
	if configSet.Get("config-url") != nil {
		config := configSet.Get("config-url").Value
		if config != nil {
			configValue := config.(string)
			if strings.HasPrefix(configValue, "http://") {
				if configSet.Get("ca-cert") != nil {
					caCert := configSet.Get("ca-cert").Value
					if caCert != nil {
						SetTransport(caCert.(string))
					}
				} else {
					SetTransport("")
				}

				kclient := khttp.NewHTTPClient()
				res, err := kclient.Get(configValue, nil)
				if err == nil {
					appConfiguration := &kconnectv1alpha.Configuration{}
					if err := json.Unmarshal([]byte(res.Body()), appConfiguration); err == nil {
						oidc := appConfiguration.Spec.Providers["oidc"]
						for k, v := range oidc {
							if k != "" && v != "" {
								addItem(configSet, k, v)
							}
						}
					} else {
						p.logger.Errorf("Error loading payload from config URL, error is: %w", err)
					}
				} else {
					p.logger.Errorf("Error calling config URL, error is: %w", err)
				}
			}
		}
	}

}

func addItem(configSet config.ConfigurationSet, key string, value string) {
	if configSet.Exists(key) {
		configSet.SetValue(key, value)
	} else {
		configSet.Add(
			&config.Item{Name: key, Type: config.ItemType("string"), Value: value, DefaultValue: ""})
	}
}

func SetTransport(file string) {

	var config *tls.Config
	if file != "" {
		caCert, err := os.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		config = &tls.Config{
			RootCAs: caCertPool,
		}
	} else {
		config = &tls.Config{InsecureSkipVerify: true}
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = config
}

// ConfigurationItems will return the configuration items for the identity plugin based
// of the cluster provider that its being used in conjunction with
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := oidc.SharedConfig()
	return cs, nil
}
