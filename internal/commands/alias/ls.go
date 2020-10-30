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

var (
	shortDescLs = "List all the aliases currently defined"
	longDescLs  = `List all the aliases currently defined for connection history entries in the
user's connection history.

An alias is a user-friendly name for a connection history entry.
`
	examplesLs = `  # Display all the aliases as a table
  kconnect alias ls

  # Display all connection history entry aliases as a table
  kconnect alias ls

  # Display all connection history entry aliases as json
  kconnect alias ls --output json

  # Connect to a cluster using a connection history entry alias
  kconnect to ${alias}

  # List all connection history entries as a table - includes aliases
  kconnect ls
`
)

func lsCommand() (*cobra.Command, error) { //nolint: dupl
	cfg := config.NewConfigurationSet()

	lsCmd := &cobra.Command{
		Use:     "ls",
		Short:   shortDescLs,
		Long:    longDescLs,
		Example: examplesLs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)

			if err := config.ApplyToConfigSet(cfg); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Debug("running `alias ls` command")
			params := &app.AliasListInput{}

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

			return a.AliasList(ctx, params)
		},
	}

	if err := addConfigLs(cfg); err != nil {
		return nil, fmt.Errorf("add ls command config: %w", err)
	}

	if err := flags.CreateCommandFlags(lsCmd, cfg); err != nil {
		return nil, err
	}

	return lsCmd, nil

}

func addConfigLs(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if _, err := cs.String("output", "table", "Output format for the results"); err != nil {
		return fmt.Errorf("adding output config item: %w", err)
	}

	cs.SetHistoryIgnore("output") //nolint

	return nil
}
