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

package history //nolint: dupl

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

const (
	shortDescImport = "Import history from an external file"
	longDescImport  = `
Allows users to import history from an external file. This can then be viewed
using the ls command, or connected to using the to command.

These imported entries will be merged with existing ones. If there is a
conflict, then the conflicting entry will not be imported (unless
--overwrite flag is supplied).

Users can optionally set any fields in the imported entry
`
	examplesImport = `
  # Imports the file into your history
  {{.CommandPath}} history import -f importfile.yaml

  # Overwrite conflicting entries
  {{.CommandPath}} history import -f importfile.yaml --overwrite

  # Wipe existing history
  {{.CommandPath}} history import -f importfile.yaml --clean

  # Set username and namespace for imported entries
  {{.CommandPath}} history import -f importfile.yaml --set username=MYUSER,namespace=kube-system

  # Only import entries that match filter
  {{.CommandPath}} history import -f importfile.yaml --filter region=us-east-1,alias=*dev*
`
)

func importCommand() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	importCmd := &cobra.Command{
		Use:     "import",
		Short:   shortDescImport,
		Long:    longDescImport,
		Example: examplesImport,
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
			zap.S().Debug("running `history import` command")
			params := &app.HistoryImportInput{}

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

			return a.HistoryImport(cmd.Context(), params)
		},
	}
	utils.FormatCommand(importCmd)

	if err := addConfigImport(cfg); err != nil {
		return nil, fmt.Errorf("adding import command config: %w", err)
	}

	if err := flags.CreateCommandFlags(importCmd, cfg); err != nil {
		return nil, err
	}

	return importCmd, nil

}

func addConfigImport(cs config.ConfigurationSet) error {

	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location config items: %w", err)
	}

	if err := app.AddHistoryImportConfig(cs); err != nil {
		return fmt.Errorf("adding history import config items: %w", err)
	}
	return nil
}
