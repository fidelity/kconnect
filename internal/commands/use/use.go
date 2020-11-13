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

package use

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/provider"
)

var (
	ErrMissingProvider       = errors.New("required provider name argument")
	ErrMissingIdpProtocol    = errors.New("missing idp protocol, please use --idp-protocol")
	ErrMustBeDirectory       = errors.New("specified config directory is not a directory")
	ErrUnsuportedIdpProtocol = errors.New("unsupported idp protocol")
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
The use command requires a target provider name as its first parameter.
`
	eksDescNote = `
* Note: kconnect use eks requires aws-iam-authenticator.
  [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)

`
	usageExample = `
  # Connect to EKS and choose an available EKS cluster.
  kconnect use eks

  # Connect to an EKS cluster and create an alias for its connection history entry.
  kconnect use eks --alias mycluster
`
	usageExampleFoot = `
  # Reconnect to a cluster by its connection history entry alias.
  kconnect to mycluster

  # Display the user's connection history as a table.
  kconnect ls
`
)

// Command creates the use command
func Command() (*cobra.Command, error) {
	longDesc := longDescHead + longDescBody + longDescFoot + eksDescNote
	useCmd := &cobra.Command{
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

	// Add the provider subcommands
	for _, provider := range provider.ListClusterProviders() {
		providerCmd, err := createProviderCmd(provider)
		if err != nil {
			return nil, fmt.Errorf("creating provider command for %s: %w", provider.Name(), err)
		}
		useCmd.AddCommand(providerCmd)
	}

	return useCmd, nil
}

func createProviderCmd(clusterProvider provider.ClusterProvider) (*cobra.Command, error) {
	params := &app.UseParams{
		Context:     provider.NewContext(),
		Provider:    clusterProvider,
		IgnoreAlias: false,
	}

	providerLongDesc := fmt.Sprintf(longDescProviderHead, clusterProvider.Name()) + longDescBody
	if clusterProvider.Name() == "eks" {
		providerLongDesc += eksDescNote
	}
	providerUsageExample := clusterProvider.UsageExample() + usageExampleFoot

	providerCmd := &cobra.Command{
		Use:     clusterProvider.Name(),
		Short:   fmt.Sprintf(shortDescProvider, clusterProvider.Name()),
		Long:    providerLongDesc,
		Example: providerUsageExample,
		AdditionalSetupE: func(cmd *cobra.Command, args []string) error {
			if err := setupIdpProtocol(os.Args, params); err != nil {
				return fmt.Errorf("additional command setup: %w", err)
			}

			if err := flags.CreateCommandFlags(cmd, params.Context.ConfigurationItems()); err != nil {
				return err
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, params.Context.ConfigurationItems())
			if err := config.ApplyToConfigSetWithProvider(params.Context.ConfigurationItems(), clusterProvider.Name()); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			if err := config.Unmarshall(params.Context.ConfigurationItems(), params); err != nil {
				return fmt.Errorf("unmarshalling config into use params: %w", err)
			}

			if params.IdentityProvider == nil {
				return ErrMissingIdpProtocol
			}

			return preRun(params)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Infow("running `use` command", "provider", clusterProvider.Name())

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

			a := app.New(app.WithHistoryStore(store))

			return a.Use(params)
		},
	}

	if err := addConfig(params.Context.ConfigurationItems(), params.Provider); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(providerCmd, params.Context.ConfigurationItems()); err != nil {
		return nil, err
	}

	providerCmd.SetUsageFunc(providerUsage(clusterProvider.Name()))

	return providerCmd, nil
}

func addConfig(cs config.ConfigurationSet, clusterProvider provider.ClusterProvider) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config %s: %w", clusterProvider.Name(), err)
	}
	if err := cs.AddSet(clusterProvider.ConfigurationItems()); err != nil {
		return fmt.Errorf("adding cluster provider config %s: %w", clusterProvider.Name(), err)
	}
	if _, err := cs.String("idp-protocol", "", "The idp protocol to use (e.g. saml, aad). See flags additional flags for the protocol."); err != nil {
		return fmt.Errorf("adding idp-protocol config: %w", err)
	}
	if _, err := cs.Bool("set-current", true, "Sets the current context in the kubeconfig to the selected cluster"); err != nil {
		return fmt.Errorf("adding set-current config: %w", err)
	}
	if err := provider.AddCommonIdentityConfig(cs); err != nil {
		return fmt.Errorf("adding common identity config items: %w", err)
	}
	if err := provider.AddCommonClusterConfig(cs); err != nil {
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

	cs.SetHistoryIgnore("set-current") //nolint

	return nil
}

func setupIdpProtocol(args []string, params *app.UseParams) error {
	idpProtocol, err := getIdpProtocol(args, params)
	if err != nil {
		return fmt.Errorf("getting idp-protocol: %w", err)
	}
	if idpProtocol == "" {
		return nil
	}

	if !isIdpSupported(idpProtocol, params.Provider) {
		zap.S().Warnw("Unsupported IdP protocol", "allowed", params.Provider.SupportedIDs(), "used", idpProtocol)
		return fmt.Errorf("using identity provider %s: %w", idpProtocol, ErrUnsuportedIdpProtocol)
	}

	idProvider, err := provider.GetIdentityProvider(idpProtocol)
	if err != nil {
		return fmt.Errorf("creating identity provider %s: %w", idpProtocol, err)
	}
	if err := params.Context.ConfigurationItems().SetValue("idp-protocol", idpProtocol); err != nil {
		return fmt.Errorf("setting idp-protocol value: %w", err)
	}
	params.IdentityProvider = idProvider

	idProviderCfg, _ := idProvider.ConfigurationItems(params.Provider.Name())
	if err := params.Context.ConfigurationItems().AddSet(idProviderCfg); err != nil {
		return err
	}

	return nil
}

func preRun(params *app.UseParams) error {
	// Update the context now the flags have been parsed
	interactive := isInteractive(params.Context.ConfigurationItems())

	params.Context = provider.NewContext(
		provider.WithInteractive(interactive),
		provider.WithConfig(params.Context.ConfigurationItems()),
	)

	identity, err := params.IdentityProvider.Authenticate(params.Context, params.Provider.Name())
	if err != nil {
		return fmt.Errorf("authenticating using provider %s: %w", params.IdentityProvider.Name(), err)
	}
	params.Identity = identity

	return params.Provider.ConfigurationResolver().Resolve(params.Context.ConfigurationItems(), params.Identity)
}

func getIdpProtocol(args []string, params *app.UseParams) (string, error) {
	// look for a flag first
	for i, arg := range args {
		if arg == "--idp-protocol" {
			return args[i+1], nil
		}
	}

	// look in app config
	return config.GetValue("idp-protocol", params.Provider.Name())
}

func isIdpSupported(idProviderName string, clusterprovider provider.ClusterProvider) bool {
	supportedIDProviders := clusterprovider.SupportedIDs()

	for _, supportedIDProvider := range supportedIDProviders {
		if supportedIDProvider == idProviderName {
			return true
		}
	}

	return false
}

func isInteractive(cs config.ConfigurationSet) bool {
	item := cs.Get("non-interactive")
	if item == nil {
		return true
	}

	value := item.Value.(bool)

	return !value
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
		usage := []string{fmt.Sprintf("Usage: %s", cmd.UseLine())}

		if cmd.Example != "" {
			usage = append(usage, "\nExamples:")
			usage = append(usage, cmd.Example)
		}

		usage = append(usage, "\nFlags:")
		usage = append(usage, cmd.LocalFlags().FlagUsages())

		usage = append(usage, "\nGlobal Flags:")
		usage = append(usage, cmd.InheritedFlags().FlagUsages())

		clusterProvider, err := provider.GetClusterProvider(providerName)
		if err != nil {
			return err
		}

		for _, idProviderName := range clusterProvider.SupportedIDs() {
			idProvider, err := provider.GetIdentityProvider(idProviderName)
			if err != nil {
				return err
			}
			usage = append(usage, fmt.Sprintf("\n%s Flags:", strings.ToUpper(idProvider.Name())))
			usage = append(usage, fmt.Sprintf("(use --idp-protocol=%s)\n", idProvider.Name()))

			cfg, err := idProvider.ConfigurationItems(providerName)
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
