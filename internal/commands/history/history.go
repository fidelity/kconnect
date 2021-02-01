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

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/utils"
)

const (
	maxHistoryEntries = 100
	shortDesc         = "Import and export history"
	longDesc          = `
Command to allow users to import or export history files.

A common use case would be for one member of a team to generate the history + 
alias config for their teams cluster(s). They could then send this file out to
the rest of the team, who can then import it. On import, they can set their
username for all of the history entries.
`
	examples = `
	# Export all history entries that have alias = *dev*
	{{.CommandPath}} history export -f history.yaml --filter alias=*dev*

	# Import history entries and set username
	{{.CommandPath}} history import -f history.yaml --set username=myuser
`
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	historyCmd := &cobra.Command{
		Use:     "history",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				zap.S().Debugw("ingoring cobra error",
					"error",
					err.Error())
			}
		},
	}
	utils.FormatCommand(historyCmd)

	if err := addConfigRoot(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	commonFlags, err := flags.CreateFlagsFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating common command flags: %w", err)
	}
	historyCmd.PersistentFlags().AddFlagSet(commonFlags)

	importCmd, err := importCommand()
	if err != nil {
		return nil, fmt.Errorf("creating history import command: %w", err)
	}
	historyCmd.AddCommand(importCmd)

	exportCmd, err := exportCommand()
	if err != nil {
		return nil, fmt.Errorf("creating history export command: %w", err)
	}
	historyCmd.AddCommand(exportCmd)

	return historyCmd, nil

}

func addConfigRoot(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location config items: %w", err)
	}
	return nil
}
