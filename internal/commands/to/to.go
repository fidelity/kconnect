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
	ErrAliasOrID       = errors.New("cannot specify id and alias at the same time")
)

func Command() (*cobra.Command, error) {
	logger := logrus.New().WithField("command", "to")

	params := &app.ConnectToParams{
		Context: provider.NewContext(provider.WithLogger(logger)),
	}

	toCmd := &cobra.Command{
		Use:   "to",
		Short: "re-connect to a previously connected cluster using your history",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.PopulateConfigFromFlags(cmd.Flags(), params.Context.ConfigurationItems())
			if err := config.Unmarshall(params.Context.ConfigurationItems(), params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}

			if params.Alias == "" && params.EntryID == "" {
				return ErrAliasIDRequired
			}

			if params.Alias != "" && params.EntryID != "" {
				return ErrAliasOrID
			}

			params.Context = provider.NewContext(
				provider.WithLogger(logger),
				provider.WithConfig(params.Context.ConfigurationItems()),
			)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			historyLoader, err := loader.NewFileLoader(params.HistoryLocation)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.HistoryLocation, err)
			}
			store, err := history.NewStore(params.HistoryMaxItems, historyLoader)
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

	flagsToAdd, err := flags.CreateFlagsFromConfig(params.Context.ConfigurationItems())
	if err != nil {
		return nil, fmt.Errorf("creating flags: %w", err)
	}
	toCmd.Flags().AddFlagSet(flagsToAdd)

	return toCmd, nil
}

func addConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("alias", "", "alias name for a history entry to connect to"); err != nil {
		return fmt.Errorf("adding alias config: %w", err)
	}
	if _, err := cs.String("id", "", "id for a history entry to connect to"); err != nil {
		return fmt.Errorf("adding id config: %w", err)
	}
	if err := app.AddHistoryConfigItems(cs); err != nil {
		return fmt.Errorf("adding history config items: %w", err)
	}
	if err := app.AddKubeconfigConfigItems(cs); err != nil {
		return fmt.Errorf("adding kubeconfig config items: %w", err)
	}

	return nil
}
