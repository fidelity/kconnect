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

package ls

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/helpers"
	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/utils"
)

var (
	shortDesc = "Query the user's connection history"
	longDesc  = `
Query and display the user's connection history entries, including entry IDs and
aliases.

Each time kconnect creates a new kubectl context to connect to a Kubernetes
cluster, it saves the settings for the new connection as an entry in the user's
connection history.  The user can then reconnect using those same settings later
via the connection history entry's ID or alias.
`
	examples = `
  # Display all connection history entries as a table
  {{.CommandPath}} ls

  # Display all connection history entries as YAML
  {{.CommandPath}} ls --output yaml

  # Display a specific connection history entry by entry id
  {{.CommandPath}} ls --filter id=01EM615GB2YX3C6WZ9MCWBDWBF

  # Display a specific connection history entry by its alias
  {{.CommandPath}} ls --filter alias=mydev

  # Display all connection history entries that have "dev" in its alias
  {{.CommandPath}} ls --filter alias=*dev*

  # Display all connection history entries for the EKS managed cluster provider
  {{.CommandPath}} ls --filter cluster-provider=eks

  # Display all connection history entries for entries with namespace kube-system
  {{.CommandPath}} ls --filter namespace=kube-system

  # Reconnect using the connection history entry alias
  {{.CommandPath}} to mydev
`
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	lsCmd := &cobra.Command{
		Use:     "ls",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)
			commonCfg, err := helpers.GetCommonConfig(cmd, cfg)
			if err != nil {
				return fmt.Errorf("gettng common config: %w", err)
			}
			if err := config.ApplyToConfigSet(commonCfg.ConfigFile, cfg); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Debug("running `ls` command")
			params := &app.HistoryQueryInput{}

			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
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

			return a.QueryHistory(cmd.Context(), params)
		},
	}
	utils.FormatCommand(lsCmd)

	if err := addConfig(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(lsCmd, cfg); err != nil {
		return nil, err
	}

	return lsCmd, nil

}

func addConfig(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryQueryConfig(cs); err != nil {
		return fmt.Errorf("adding history query config items: %w", err)
	}
	if err := app.AddHistoryConfigItems(cs); err != nil {
		return fmt.Errorf("adding history config items: %w", err)
	}

	if err := app.AddKubeconfigConfigItems(cs); err != nil {
		return fmt.Errorf("adding kubeconfig config items: %w", err)
	}
	return nil
}
