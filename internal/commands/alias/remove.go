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

	"github.com/fidelity/kconnect/internal/helpers"
	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/utils"
)

var (
	shortDescRemove = "Remove connection history entry aliases."
	longDescRemove  = `
Remove an alias from a single connection history entry by the entry ID or the
alias.

Set the --all flag on this command to remove all connection history aliases from
the user's connection history.
`
	examplesRemove = `
  # Remove an alias using the alias name
  {{.CommandPath}} alias remove --alias dev-bu-1

  # Remove an alias using a histiry entry id
  {{.CommandPath}} alias remove --id 01EMEM5DB60TMX7D8SS2JCX3MT

  # Remove all aliases
  {{.CommandPath}} alias remove --all

  # List available aliases
  {{.CommandPath}} alias ls

  # Query your connection history - includes aliases
  {{.CommandPath}} ls
`
)

func removeCommand() (*cobra.Command, error) { //nolint: dupl
	cfg := config.NewConfigurationSet()

	rmCmd := &cobra.Command{
		Use:     "remove",
		Short:   shortDescRemove,
		Long:    longDescRemove,
		Example: examplesRemove,
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
			zap.S().Debug("running `alias remove` command")
			params := &app.AliasRemoveInput{}

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

			return a.AliasRemove(cmd.Context(), params)
		},
	}
	utils.FormatCommand(rmCmd)

	if err := addConfigRemove(cfg); err != nil {
		return nil, fmt.Errorf("add remove command config: %w", err)
	}

	if err := flags.CreateCommandFlags(rmCmd, cfg); err != nil {
		return nil, err
	}

	return rmCmd, nil

}

func addConfigRemove(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location config: %w", err)
	}
	if err := app.AddHistoryIdentifierConfig(cs); err != nil {
		return fmt.Errorf("adding history identifier config: %w", err)
	}

	if _, err := cs.Bool("all", false, "Remove all aliases from the histiry entries"); err != nil {
		return fmt.Errorf("adding all config item: %w", err)
	}

	return nil
}
