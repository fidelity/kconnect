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

package ls

import (
	"fmt"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Command() (*cobra.Command, error) {
	logger := logrus.New().WithField("command", "ls")

	cfg := config.NewConfigurationSet()

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "Query your connection history",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &app.HistoryQueryInput{}

			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)
			if err := config.ApplyToConfigSet(cfg); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}
			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}

			historyLoader, err := loader.NewFileLoader(params.Location)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.Location, err)
			}
			store, err := history.NewStore(params.MaxItems, historyLoader)
			if err != nil {
				return fmt.Errorf("creating history store: %w", err)
			}

			a := app.New(app.WithLogger(logger), app.WithHistoryStore(store))

			ctx := provider.NewContext(
				provider.WithLogger(logger),
				provider.WithConfig(cfg),
			)

			return a.QueryHistory(ctx, params)
		},
	}

	if err := addConfig(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(lsCmd, cfg); err != nil {
		return nil, err
	}

	return lsCmd, nil

}

func addConfig(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if err := app.AddHistoryQueryConfig(cs); err != nil {
		return fmt.Errorf("adding history query config items: %w", err)
	}
	if err := app.AddHistoryConfigItems(cs); err != nil {
		return fmt.Errorf("adding history config items: %w", err)
	}
	if err := app.AddKubeconfigConfigItems(cs); err != nil {
		return fmt.Errorf("adding kubeconfig config items: %w", err)
	}
	if _, err := cs.String("output", "table", "Output format for the results"); err != nil {
		return fmt.Errorf("adding output config item: %w", err)
	}

	return nil
}
