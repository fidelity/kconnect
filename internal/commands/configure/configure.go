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

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Command() (*cobra.Command, error) {
	logger := logrus.New().WithField("command", "configure")

	cfg := config.NewConfigurationSet()

	cfgCmd := &cobra.Command{
		Use:   "configure",
		Short: "Set and view your default kconnect configuration. If no flags are supplied your config is displayed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &app.ConfigureInput{}

			flags.BindFlags(cmd)
			flags.PopulateConfigFromFlags(cmd.Flags(), cfg)
			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}

			a := app.New(app.WithLogger(logger))

			ctx := provider.NewContext(
				provider.WithLogger(logger),
				provider.WithConfig(cfg),
			)

			return a.Configuration(ctx, params)
		},
	}

	if err := addConfig(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	flagsToAdd, err := flags.CreateFlagsFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating flags: %w", err)
	}
	cfgCmd.Flags().AddFlagSet(flagsToAdd)

	return cfgCmd, nil

}

func addConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("source", "", "File or remote location to use to set the default configuration"); err != nil {
		return fmt.Errorf("adding source config item: %w", err)
	}
	if _, err := cs.String("output", "yaml", "Controls the output format for the result."); err != nil {
		return fmt.Errorf("adding output config item: %w", err)
	}
	if err := cs.SetShort("source", "s"); err != nil {
		return fmt.Errorf("setting shorthand for source config item: %w", err)
	}

	return nil
}
