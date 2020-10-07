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

package configure

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/provider"
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	cfgCmd := &cobra.Command{
		Use:   "configure",
		Short: "Set and view your default kconnect configuration. If no flags are supplied your config is displayed.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)

			if err := config.ApplyToConfigSet(cfg); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &app.ConfigureInput{}

			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}

			a := app.New()

			ctx := provider.NewContext(
				provider.WithConfig(cfg),
			)

			return a.Configuration(ctx, params)
		},
	}

	if err := addConfig(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(cfgCmd, cfg); err != nil {
		return nil, err
	}

	return cfgCmd, nil

}

func addConfig(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if _, err := cs.String("file", "", "File or remote location to use to set the default configuration"); err != nil {
		return fmt.Errorf("adding file config item: %w", err)
	}
	if _, err := cs.String("output", "yaml", "Controls the output format for the result."); err != nil {
		return fmt.Errorf("adding output config item: %w", err)
	}
	if err := cs.SetShort("file", "f"); err != nil {
		return fmt.Errorf("setting shorthand for file config item: %w", err)
	}

	cs.SetHistoryIgnore("file")   //nolint
	cs.SetHistoryIgnore("output") //nolint

	return nil
}
