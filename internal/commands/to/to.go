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

package to

import (
	"errors"
	"fmt"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ErrAliasIDRequired = errors.New("alias or id must be specified")
)

func Command() (*cobra.Command, error) {
	logger := logrus.New().WithField("command", "to")

	params := &app.ConnectToParams{
		Context: provider.NewContext(provider.WithLogger(logger)),
	}

	toCmd := &cobra.Command{
		Use:   "to [historyid/alias]",
		Short: "Re-connect to a previously connected cluster using your history",
		Long: `use is for re-connecting to a previously connected cluster using your history.
You can use the history id or alias as the argument.`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return ErrAliasIDRequired
			}
			params.AliasOrID = args[0]

			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, params.Context.ConfigurationItems())

			if err := config.Unmarshall(params.Context.ConfigurationItems(), params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}

			params.Context = provider.NewContext(
				provider.WithLogger(logger),
				provider.WithConfig(params.Context.ConfigurationItems()),
			)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			historyLoader, err := loader.NewFileLoader(params.Location)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.Location, err)
			}
			store, err := history.NewStore(params.MaxItems, historyLoader)
			if err != nil {
				return fmt.Errorf("creating history store: %w", err)
			}

			a := app.New(app.WithLogger(logger), app.WithHistoryStore(store))

			return a.ConnectTo(params)
		},
	}

	if err := addConfig(params.Context.ConfigurationItems()); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(toCmd, params.Context.ConfigurationItems()); err != nil {
		return nil, err
	}

	return toCmd, nil
}

func addConfig(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if _, err := cs.Bool("set-current", true, "Sets the current context in the kubeconfig to the selected cluster"); err != nil {
		return fmt.Errorf("adding set-current config: %w", err)
	}
	if _, err := cs.String("password", "", "Password to use"); err != nil {
		return fmt.Errorf("adding password config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location items: %w", err)
	}
	if err := app.AddKubeconfigConfigItems(cs); err != nil {
		return fmt.Errorf("adding kubeconfig config items: %w", err)
	}

	return nil
}
