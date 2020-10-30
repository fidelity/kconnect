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
)

const (
	maxHistoryEntries = 100
	shortDesc         = "Query and manipulate connection history entry aliases."
	longDesc          = `
An alias is a user-friendly name for a connection history entry, otherwise 
referred to by its entry ID. 

The alias command and sub-commands allow you to query and manipulate aliases for
connection history entries.
`
	examples = `
  # Add an alias to an existing connection history entry
  kconnect alias add --id 123456 --alias appdev

  # List available connection history entry aliases
  kconnect alias ls

  # Remove an alias from a connection history entry
  kconnect alias remove --alias appdev
`
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	aliasCmd := &cobra.Command{
		Use:     "alias",
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

	if err := addConfigRoot(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	commonFlags, err := flags.CreateFlagsFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating common command flags: %w", err)
	}
	aliasCmd.PersistentFlags().AddFlagSet(commonFlags)

	lsCmd, err := lsCommand()
	if err != nil {
		return nil, fmt.Errorf("creating alias ls command: %w", err)
	}
	aliasCmd.AddCommand(lsCmd)

	addCmd, err := addCommand()
	if err != nil {
		return nil, fmt.Errorf("creating alias add command: %w", err)
	}
	aliasCmd.AddCommand(addCmd)

	removeCmd, err := removeCommand()
	if err != nil {
		return nil, fmt.Errorf("creating alias remove command: %w", err)
	}
	aliasCmd.AddCommand(removeCmd)

	return aliasCmd, nil

}

func addConfigRoot(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history config items: %w", err)
	}

	return nil
}
