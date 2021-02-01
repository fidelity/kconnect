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

	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/utils"
)

const (
	shortDescExport = "Export history to an external file"
	longDescExport  = `
Allows users to export history to an external file. This file can then be
imported by another user using the import command"
`
	examplesExport = `
  # Export your history into a file
  {{.CommandPath}} history export -f exportfile.yaml

  # Set username and namespace for exported entries
  {{.CommandPath}} history export -f exportfile.yaml --set username=MYUSER,namespace=kube-system

  # Only export entries that match filter
  {{.CommandPath}} history export -f exportfile.yaml --filter region=us-east-1,alias=*dev*
`
)

func exportCommand() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	importCmd := &cobra.Command{
		Use:     "export",
		Short:   shortDescExport,
		Long:    longDescExport,
		Example: examplesExport,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)

			if err := config.ApplyToConfigSet(cfg); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Debug("running `history export` command")
			params := &app.HistoryExportInput{}

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

			return a.HistoryExport(cmd.Context(), params)
		},
	}
	utils.FormatCommand(importCmd)

	if err := addConfigExport(cfg); err != nil {
		return nil, fmt.Errorf("adding export command config: %w", err)
	}

	if err := flags.CreateCommandFlags(importCmd, cfg); err != nil {
		return nil, err
	}

	return importCmd, nil

}

func addConfigExport(cs config.ConfigurationSet) error {

	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location config items: %w", err)
	}

	if err := app.AddHistoryExportConfig(cs); err != nil {
		return fmt.Errorf("adding history export config items: %w", err)
	}
	return nil
}
