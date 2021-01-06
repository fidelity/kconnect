/*
Copyright 2021 The kconnect Authors.

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

package util

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
)

const (
	shortDesc = "Utility commands."

	examples = `
  # Import history from a file
  kconnect util import-history -h history.yaml

  # Import history from a file and overwrite any existing history
  kconnect util import-history -h history.yaml --overwrite
`
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	utilsCmd := &cobra.Command{
		Use:     "util",
		Short:   shortDesc,
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
	utilsCmd.PersistentFlags().AddFlagSet(commonFlags)

	//TODO: add subcommands
	// lsCmd, err := lsCommand()
	// if err != nil {
	// 	return nil, fmt.Errorf("creating alias ls command: %w", err)
	// }
	// utilsCmd.AddCommand(lsCmd)

	return utilsCmd, nil

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
