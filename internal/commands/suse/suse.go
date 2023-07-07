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

package suse

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	kconnectv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/internal/helpers"
	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/defaults"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/registry"
	"github.com/fidelity/kconnect/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ErrMissingProvider    = errors.New("required provider name argument")
	ErrMissingIdpProtocol = errors.New("missing idp protocol")
	ErrMustBeDirectory    = errors.New("specified config directory is not a directory")
)

const (
	shortDesc         = "Connect to a Kubernetes cluster provider and cluster."
	shortDescProvider = "Connect to the %s cluster provider and choose a cluster."
	longDescHead      = `
Connect to a managed Kubernetes cluster provider via the configured identity
provider, prompting the user to enter or choose connection settings appropriate
to the provider and a target cluster once connected.
`
	longDescProviderHead = `
Connect to %s via the configured identify provider, prompting the user to enter
or choose connection settings and a target cluster once connected.
`
	longDescBody = `
The kconnect tool generates a kubectl configuration context with a fresh access
token to connect to the chosen cluster and adds a connection history entry to
store the chosen connection settings.  If given an alias name, kconnect will add
a user-friendly alias to the new connection history entry.

The user can then reconnect to the provider with the settings stored in the
connection history entry using the kconnect to command and the connection history
entry ID or alias.  When the user reconnects using a connection history entry,
kconnect regenerates the kubectl configuration context and refreshes their access
token.
`
	longDescFoot = `
The use command requires a target provider name as its first parameter. If no
value is supplied for --idp-protocol the first supported protocol for the
specified cluster provider.

* Note: interactive mode is not supported in windows git-bash application currently.
`
	eksDescNote = `
* Note: kconnect use eks requires aws-iam-authenticator.
  [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)
`
	aksDescNote = `
* Note: kconnect use aks requires kubelogin and azure cli.
  [kubelogin](https://github.com/Azure/kubelogin)
  [azure-cli](https://github.com/Azure/azure-cli)
`
	usageExample = `
  # Connect to EKS and choose an available EKS cluster.
  {{.CommandPath}} use eks

  # Connect to an EKS cluster and create an alias for its connection history entry.
  {{.CommandPath}} use eks --alias mycluster
`
	usageExampleFoot = `
  # Reconnect to a cluster by its connection history entry alias.
  {{.CommandPath}} to mycluster

  # Display the user's connection history as a table.
  {{.CommandPath}} ls
`
)

// Command creates the use command
func Command() (*cobra.Command, error) {
	longDesc := longDescHead + longDescBody + longDescFoot + eksDescNote + aksDescNote
	sUseCmd := &cobra.Command{
		Use:     "use",
		Short:   shortDesc,
		Long:    longDesc,
		Example: usageExample + usageExampleFoot,
		Run: func(c *cobra.Command, _ []string) {
			if err := c.Help(); err != nil {
				zap.S().Debugw("ignoring cobra error", "error", err.Error())
			}
		},
	}

	utils.FormatCommand(sUseCmd)

	// Add the provider subcommands
	for _, registration := range registry.ListDiscoveryPluginRegistrations() {
		providerCmd, err := createProviderCmd(registration)
		if err != nil {
			return nil, fmt.Errorf("creating provider command for %s: %w", registration.Name, err)
		}
		sUseCmd.AddCommand(providerCmd)
	}

	return sUseCmd, nil
}

func createProviderCmd(registration *registry.DiscoveryPluginRegistration) (*cobra.Command, error) {
	params := &app.UseInput{
		IgnoreAlias:       false,
		ConfigSet:         config.NewConfigurationSet(),
		DiscoveryProvider: registration.Name,
	}

	providerLongDesc := fmt.Sprintf(longDescProviderHead, registration.Name) + longDescBody
	if registration.Name == "eks" {
		providerLongDesc += eksDescNote
	} else if registration.Name == "aks" {
		providerLongDesc += aksDescNote
	}
	providerUsageExample := registration.UsageExample + usageExampleFoot

	providerCmd := &cobra.Command{
		Use:     registration.Name,
		Short:   fmt.Sprintf(shortDescProvider, registration.Name),
		Long:    providerLongDesc,
		Example: providerUsageExample,
		AdditionalSetupE: func(cmd *cobra.Command, args []string) error {
			if err := setupIdpProtocol(cmd, os.Args, params); err != nil {
				return fmt.Errorf("additional command setup: %w", err)
			}

			if err := flags.CreateCommandFlags(cmd, params.ConfigSet); err != nil {
				return err
			}

			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, params.ConfigSet)
			commonCfg, err := helpers.GetCommonConfig(cmd, params.ConfigSet)
			if err != nil {
				return fmt.Errorf("gettng common config: %w", err)
			}
			if err := config.ApplyToConfigSetWithProvider(commonCfg.ConfigFile, params.ConfigSet, registration.Name); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			if err := config.Unmarshall(params.ConfigSet, params); err != nil {
				return fmt.Errorf("unmarshalling config into use params: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Debugw("running `use` command", "provider", registration.Name)
			setupIdpProtocol(cmd, os.Args, params)
			if err := ensureConfigFolder(defaults.AppDirectory()); err != nil {
				return fmt.Errorf("ensuring app directory exists: %w", err)
			}

			historyLoader, err := loader.NewFileLoader(params.Location)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.Location, err)
			}
			store, err := history.NewStore(params.MaxItems, historyLoader)
			if err != nil {
				return fmt.Errorf("creating history store: %w", err)
			}

			a := app.New(app.WithHistoryStore(store), app.WithInteractive(!params.NoInput))

			return a.Suse(cmd.Context(), params)
		},
	}

	if err := addConfig(params.ConfigSet, registration); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(providerCmd, params.ConfigSet); err != nil {
		return nil, err
	}

	providerCmd.SetUsageFunc(providerUsage(registration.Name))

	utils.FormatCommand(providerCmd)
	return providerCmd, nil
}

func addConfig(cs config.ConfigurationSet, registration *registry.DiscoveryPluginRegistration) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config %s: %w", registration.Name, err)
	}
	providerCS, err := registration.ConfigurationItemsFunc("")
	if err != nil {
		return fmt.Errorf("getting configuration items for %s: %w", registration.Name, err)
	}
	if err := cs.AddSet(providerCS); err != nil {
		return fmt.Errorf("adding cluster provider config %s: %w", registration.Name, err)
	}
	if _, err := cs.String("idp-protocol", "", "The idp protocol to use (e.g. saml, aad). See flags additional flags for the protocol."); err != nil {
		return fmt.Errorf("adding idp-protocol config: %w", err)
	}
	if _, err := cs.Bool("set-current", true, "Sets the current context in the kubeconfig to the selected cluster"); err != nil {
		return fmt.Errorf("adding set-current config: %w", err)
	}
	if err := common.AddCommonIdentityConfig(cs); err != nil {
		return fmt.Errorf("adding common identity config items: %w", err)
	}
	if err := common.AddCommonClusterConfig(cs); err != nil {
		return fmt.Errorf("adding common cluster config items: %w", err)
	}
	if err := app.AddHistoryConfigItems(cs); err != nil {
		return fmt.Errorf("adding history config items: %w", err)
	}
	if err := app.AddKubeconfigConfigItems(cs); err != nil {
		return fmt.Errorf("adding kubeconfig config items: %w", err)
	}
	if err := app.AddCommonUseConfigItems(cs); err != nil {
		return fmt.Errorf("adding common use config items: %w", err)
	}

	cs.SetHistoryIgnore("set-current") //nolint: errcheck

	return nil
}

func setupIdpProtocol(cmd *cobra.Command, args []string, params *app.UseInput) error {
	idpProtocol, _, _ := getIdpProtocol(args, params)

	if err := params.ConfigSet.SetValue("idp-protocol", idpProtocol); err != nil {
		return fmt.Errorf("setting idp-protocol value: %w", err)
	}

	return nil
}

func getIdpProtocol(args []string, params *app.UseInput) (string, bool, error) {
	// look for a flag first
	getFromConfig(params, "oidc-user")
	getFromConfig(params, "oidc-server")
	getFromConfig(params, "oidc-client-id")
	getFromConfig(params, "oidc-client-secret")
	getFromConfig(params, "oidc-tenant-id")
	getFromConfig(params, "cluster-url")
	getFromConfig(params, "cluster-auth")
	getFromConfig(params, "cluster-id")

	getFromConfig(params, "login")
	getFromConfig(params, "azure-kubelogin")
	getFromConfig(params, "config-url")
	getFromConfig(params, "ca-cert")
	getFromConfig(params, "oidc-use-pkce")

	for i, arg := range args {
		if arg == "--config-url" {
			addItem(params, "config-url", args[i+1])
		}
		if arg == "--ca-cert" {
			addItem(params, "ca-cert", args[i+1])
		}
	}

	if params.ConfigSet.Get("config-url") != nil {
		config := params.ConfigSet.Get("config-url").Value
		if config != nil {
			configValue := config.(string)
			if strings.HasPrefix(configValue, "https://") {
				if params.ConfigSet.Get("ca-cert") != nil {
					caCert := params.ConfigSet.Get("ca-cert").Value
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
						eks := appConfiguration.Spec.Providers["eks"]
						for k, v := range eks {
							if k != "" && v != "" {
								addItem(params, k, v)
							}
						}
					} else {
						panic("bad http config body")
					}
				} else {
					panic(err)
				}
			}
		}
	}

	for i, arg := range args {
		if arg == "--oidc-user" {
			addItem(params, "oidc-user", args[i+1])
		}
		if arg == "--oidc-server" {
			addItem(params, "oidc-server", args[i+1])
		}
		if arg == "--oidc-client-id" {
			addItem(params, "oidc-client-id", args[i+1])
		}
		if arg == "--oidc-client-secret" {
			addItem(params, "oidc-client-secret", args[i+1])
		}
		if arg == "--oidc-tenant-id" {
			addItem(params, "oidc-tenant-id", args[i+1])
		}
		if arg == "--cluster-url" {
			addItem(params, "cluster-url", args[i+1])
		}
		if arg == "--cluster-auth" {
			addItem(params, "cluster-auth", args[i+1])
		}
		if arg == "--cluster-id" {
			addItem(params, "cluster-id", args[i+1])
		}
		if arg == "--login" {
			addItem(params, "login", args[i+1])
		}
		if arg == "--azure-kubelogin" {
			addItem(params, "azure-kubelogin", args[i+1])
		}
		if arg == "--oidc-use-pkce" {
			addItem(params, "oidc-use-pkce", "true")
		}
	}
	for i, arg := range args {
		if arg == "--idp-protocol" {
			return args[i+1], true, nil
		}
	}
	return "", false, nil
}

func getFromConfig(params *app.UseInput, key string) {
	value, err := config.GetValue(key, "eks")
	if err == nil && value != "" {
		addItem(params, key, value)
	}
}

func SetTransport(file string) {

	var config *tls.Config
	if file != "" {
		caCert, err := ioutil.ReadFile(file)
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

func addItem(params *app.UseInput, key string, value string) {
	if params.ConfigSet.Exists(key) {
		params.ConfigSet.SetValue(key, value)
	} else {
		params.ConfigSet.Add(
			&config.Item{Name: key, Type: config.ItemType("string"), Value: value, DefaultValue: ""})
	}
}

func ensureConfigFolder(path string) error {
	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("getting details of config folder %s: %w", path, err)
	} else if err != nil && os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			return fmt.Errorf("creating config directory %s: %w", path, err)
		}

		return nil
	}

	if !info.IsDir() {
		return ErrMustBeDirectory
	}

	return nil
}

func providerUsage(providerName string) func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		use := utils.FormatUse(cmd.UseLine())
		usage := []string{fmt.Sprintf("Usage: %s", use)}

		if cmd.Example != "" {
			usage = append(usage, "\nExamples:")
			usage = append(usage, cmd.Example)
		}

		usage = append(usage, "\nFlags:")
		usage = append(usage, cmd.LocalFlags().FlagUsages())

		usage = append(usage, "\nGlobal Flags:")
		usage = append(usage, cmd.InheritedFlags().FlagUsages())

		clusterProviderReg, err := registry.GetDiscoveryProviderRegistration(providerName)
		if err != nil {
			return err
		}

		for _, idProviderName := range clusterProviderReg.SupportedIdentityProviders {
			idProviderReg, err := registry.GetIdentityProviderRegistration(idProviderName)
			if err != nil {
				return err
			}
			usage = append(usage, fmt.Sprintf("\n%s Flags:", strings.ToUpper(idProviderReg.Name)))
			usage = append(usage, fmt.Sprintf("(use --idp-protocol=%s)\n", idProviderReg.Name))

			cfg, err := idProviderReg.ConfigurationItemsFunc(providerName)
			if err != nil {
				return err
			}

			fs, err := flags.CreateFlagsFromConfig(cfg)
			if err != nil {
				return err
			}
			usage = append(usage, fs.FlagUsages())
		}

		cmd.Println(strings.Join(usage, "\n"))
		return nil
	}
}
