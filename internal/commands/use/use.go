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
	"github.com/fidelity/kconnect/internal/usage"
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
	logger := logrus.WithField("command", "use")

	params := &app.UseParams{
		Context: provider.NewContext(provider.WithLogger(logger)),
	}

	useCmd := &cobra.Command{
		Use:   "use",
		Short: "connect to a target environment and use clusters",
		Args:  cobra.ExactArgs(1),
		AdditionalSetupE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return ErrMissingProvider
			}
			return cmdSetup(cmd, args, params, logger)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.PopulateConfigFromFlags(cmd.Flags(), params.Context.ConfigurationItems())

			if err := config.Unmarshall(params.Context.ConfigurationItems(), params); err != nil {
				return fmt.Errorf("unmarshalling config into use params: %w", err)
			}

			return preRun(params, logger)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigFolder(defaults.AppDirectory()); err != nil {
				return fmt.Errorf("ensuring app directory exists: %w", err)
			}

			historyLoader, err := loader.NewFileLoader(params.HistoryLocation)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.HistoryLocation, err)
			}
			store, err := history.NewStore(params.HistoryMaxItems, historyLoader)
			if err != nil {
				return fmt.Errorf("creating history store: %w", err)
			}

			a := app.New(app.WithLogger(logger), app.WithHistoryStore(store))
			//TODO: change the below as its horrible
			a.RegisterSensitiveFlag(cmd.Flags().Lookup("password"))

			return a.Use(params)
		},
	}

	if err := addConfig(params.Context.ConfigurationItems()); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	flagsToAdd, err := flags.CreateFlagsFromConfig(params.Context.ConfigurationItems())
	if err != nil {
		return nil, fmt.Errorf("creating flags: %w", err)
	}
	useCmd.Flags().AddFlagSet(flagsToAdd)

	useCmd.SetUsageFunc(usage.Get)

	return useCmd, nil
}

func addConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("idp-protocol", "", "the idp protocol to use (e.g. saml)"); err != nil {
		return fmt.Errorf("adding idp-protocol config: %w", err)
	}
	if _, err := cs.Bool("set-current", false, "sets the current context in the kubeconfig to selected cluster"); err != nil {
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

func cmdSetup(cmd *cobra.Command, args []string, params *app.UseParams, logger *logrus.Entry) error {
	selectedProvider, err := provider.GetClusterProvider(args[0])
	if err != nil {
		return fmt.Errorf("creating cluster provider %s: %w", args[0], err)
	}
	params.Provider = selectedProvider

	logger.Infof("using cluster provider %s", params.Provider.Name())
	providerCfg := selectedProvider.ConfigurationItems()
	if err := params.Context.ConfigurationItems().AddSet(providerCfg); err != nil {
		return err
	}

	idpProtocol := findIdpProtocolFromArgs(args)
	if idpProtocol == "" {
		return ErrMissingIdpProtocol
	}
	idProvider, err := provider.GetIdentityProvider(idpProtocol)
	if err != nil {
		return fmt.Errorf("creating identity provider %s: %w", idpProtocol, err)
	}
	params.IdentityProvider = idProvider
	logger.Infof("using identity provider %s", idProvider.Name())
	idProviderCfg := idProvider.ConfigurationItems()
	if err := params.Context.ConfigurationItems().AddSet(idProviderCfg); err != nil {
		return err
	}

	providerFlags, err := flags.CreateFlagsFromConfig(params.Context.ConfigurationItems())
	if err != nil {
		return fmt.Errorf("converting provider config to flags: %w", err)
	}

	cmd.Flags().AddFlagSet(providerFlags)

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

func findIdpProtocolFromArgs(args []string) string {
	index := -1
	for i, arg := range args {
		if arg == "--idp-protocol" {
			index = i
			break
		}
	}
	if index == -1 {
		return ""
	}

	return args[index+1]
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
