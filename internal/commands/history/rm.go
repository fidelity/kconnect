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

package history

import (
	"fmt"

	"github.com/fidelity/kconnect/internal/helpers"
	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	shortDescRm = "Remove history entries"
	longDescRm  = `
Allows users to delete history entries from their history
`
	examplesRm = `
  # Display history entries
  {{.CommandPath}} ls

  # Delete a history entry with specific ID
  {{.CommandPath}} history rm 01exm3ty400w9sr28jawc8fkae

  # Delete multiple history entries with specific IDs
  {{.CommandPath}} history rm 01exm3ty400w9sr28jawc8fkae 01exm3tvw2f5snkj18rk1ngmyb

  # Delete all history entries
  {{.CommandPath}} history rm --all

  # Delete all history entries that match the filter (e.g. alias that have "prod" in them)
  {{.CommandPath}} history rm --filter alias=*prod*
`
)

func rmCommand() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	importCmd := &cobra.Command{
		Use:     "rm",
		Args:    cobra.ArbitraryArgs,
		Short:   shortDescRm,
		Long:    longDescRm,
		Example: examplesRm,
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
			zap.S().Debug("running `history rm` command")

			params := &app.HistoryRemoveInput{
				RemoveList: args,
			}

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

			return a.HistoryRemove(cmd.Context(), params)
		},
	}
	utils.FormatCommand(importCmd)

	if err := addConfigRm(cfg); err != nil {
		return nil, fmt.Errorf("adding export command config: %w", err)
	}

	if err := flags.CreateCommandFlags(importCmd, cfg); err != nil {
		return nil, err
	}

	return importCmd, nil

}

func addConfigRm(cs config.ConfigurationSet) error {

	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location config items: %w", err)
	}

	if err := app.AddHistoryRemoveConfig(cs); err != nil {
		return fmt.Errorf("adding history remove config items: %w", err)
	}
	return nil
}
