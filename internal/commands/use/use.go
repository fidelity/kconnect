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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/provider"
)

var (
	ErrMissingProvider    = errors.New("required provider name argument")
	ErrMissingIdpProtocol = errors.New("missing idp protocol, please use --idp-protocol")
	ErrMustBeDirectory    = errors.New("specified config directory is not a directory")
)

// Command creates the use command
func Command() (*cobra.Command, error) {
	useCmd := &cobra.Command{
		Use:   "use",
		Short: "Connect to a target environment and discover clusters for use",
		Run: func(c *cobra.Command, _ []string) {
			if err := c.Help(); err != nil {
				logrus.Debugf("ignoring cobra error %q", err.Error())
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
	logger := logrus.WithField("command", "use").WithField("provider", clusterProvider.Name())

	params := &app.UseParams{
		Context:  provider.NewContext(provider.WithLogger(logger)),
		Provider: clusterProvider,
	}

	providerCmd := &cobra.Command{
		Use:   clusterProvider.Name(),
		Short: fmt.Sprintf("Connect to %s and discover clusters for use", clusterProvider.Name()),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromFlags(cmd.Flags(), params.Context.ConfigurationItems())
			if err := config.ApplyToConfigSetWithProvider(params.Context.ConfigurationItems(), clusterProvider.Name()); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			if err := config.Unmarshall(params.Context.ConfigurationItems(), params); err != nil {
				return fmt.Errorf("unmarshalling config into use params: %w", err)
			}

			if params.IdentityProvider == nil {
				return ErrMissingIdpProtocol
			}

			return preRun(params, logger)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			a := app.New(app.WithLogger(logger), app.WithHistoryStore(store))

			return a.Use(params)
		},
	}

	if err := addConfig(params.Context.ConfigurationItems(), params.Provider); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := setupIdpProtocol(os.Args, params, logger); err != nil {
		return nil, fmt.Errorf("additional command setup: %w", err)
	}

	if err := flags.CreateFlagsAndMarkRequired(providerCmd, params.Context.ConfigurationItems()); err != nil {
		return nil, err
	}

	return providerCmd, nil
}

func addConfig(cs config.ConfigurationSet, clusterProvider provider.ClusterProvider) error {
	if err := cs.AddSet(clusterProvider.ConfigurationItems()); err != nil {
		return fmt.Errorf("adding cluster provider config %s: %w", clusterProvider.Name(), err)
	}
	if _, err := cs.String("idp-protocol", "", "The idp protocol to use (e.g. saml)"); err != nil {
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

	return nil
}

func setupIdpProtocol(args []string, params *app.UseParams, logger *logrus.Entry) error {
	idpProtocol, err := getIdpProtocol(args, params)
	if err != nil {
		return fmt.Errorf("getting idp-protocol: %w", err)
	}
	if idpProtocol == "" {
		return nil
	}

	idProvider, err := provider.GetIdentityProvider(idpProtocol)
	if err != nil {
		return fmt.Errorf("creating identity provider %s: %w", idpProtocol, err)
	}
	if err := params.Context.ConfigurationItems().SetValue("idp-protocol", idpProtocol); err != nil {
		return fmt.Errorf("setting idp-protocol value: %w", err)
	}
	params.IdentityProvider = idProvider

	logger.Infof("using identity provider %s", idProvider.Name())
	idProviderCfg := idProvider.ConfigurationItems()
	if err := params.Context.ConfigurationItems().AddSet(idProviderCfg); err != nil {
		return err
	}

	return nil
}

func preRun(params *app.UseParams, logger *logrus.Entry) error {
	// Update the context now the flags have been parsed
	interactive := isInteractive(params.Context.ConfigurationItems())

	params.Context = provider.NewContext(
		provider.WithLogger(logger),
		provider.WithInteractive(interactive),
		provider.WithConfig(params.Context.ConfigurationItems()),
	)

	identity, err := params.IdentityProvider.Authenticate(params.Context, params.Provider.Name())
	if err != nil {
		return fmt.Errorf("authenticating using provider %s: %w", params.IdentityProvider.Name(), err)
	}
	params.Identity = identity

	return params.Provider.ConfigurationResolver().Resolve(params.Context.ConfigurationItems())
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
