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

package alias

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	shortDescAdd = "Add an alias to a connection history entry"
	longDescAdd  = `Adds a user-friendly alias to a connection history entry.

The user can then reconnect and refresh the access token for that cluster using 
the alias instead of the connection history entry's unique ID.
`
	examplesAdd = `  # Add an alias to a connection history entry
  kconnect alias add --id 01EMEM5DB60TMX7D8SS2JCX3MT --alias dev-bu-1

  # Connect to a cluster using the alias
  kconnect to dev-bu-1

  # List available aliases
  kconnect alias ls

  # List available history entries - includes aliases
  kconnect ls
`
)

func addCommand() (*cobra.Command, error) { //nolint: dupl
	cfg := config.NewConfigurationSet()

	lsCmd := &cobra.Command{
		Use:     "add",
		Short:   shortDescAdd,
		Long:    longDescAdd,
		Example: examplesAdd,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)

			if err := config.ApplyToConfigSet(cfg); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Debug("running `alias add` command")
			params := &app.AliasAddInput{}

			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}

			historyLoader, err := loader.NewFileLoader(params.Location)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.Location, err)
			}
			store, err := history.NewStore(maxHistoryEntries, historyLoader)
			if err != nil {
				return fmt.Errorf("creating history store: %w", err)
			}

			a := app.New(app.WithHistoryStore(store))

			ctx := provider.NewContext(
				provider.WithConfig(cfg),
			)

			return a.AliasAdd(ctx, params)
		},
	}

	if err := addConfigAdd(cfg); err != nil {
		return nil, fmt.Errorf("adding add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(lsCmd, cfg); err != nil {
		return nil, err
	}

	return lsCmd, nil

}

func addConfigAdd(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location config: %w", err)
	}
	if err := app.AddHistoryIdentifierConfig(cs); err != nil {
		return fmt.Errorf("adding history identifier config: %w", err)
	}

	return nil
}
