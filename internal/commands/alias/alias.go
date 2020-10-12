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

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	aliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Query and manipulate aliases for your connection history.",
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

	return aliasCmd, nil

}

func addConfigRoot(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryConfigItems(cs); err != nil {
		return fmt.Errorf("adding history config items: %w", err)
	}

	return nil
}
