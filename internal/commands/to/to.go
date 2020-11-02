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

package to

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/provider"
)

var (
	ErrAliasIDRequired = errors.New("alias or id must be specified")

	examples = `  # Re-connect based on an alias
  kconnect to uat-bu1

  # # Re-connect based on an history id - history id can be found using kconnect ls
  kconnect to 01EM615GB2YX3C6WZ9MCWBDWBF

  # Re-connect to current cluster (this is useful for renewing credentials)
  kconnect to -
  OR 
  kconnect to LAST

  # Re-connect to cluster used before current one
  kconnect to LAST~1

  # Re-connect based on an alias supplying a password
  kconnect to uat-bu1 --password supersecret

  # Re-connect based on an alias supplying a password via env var
  KCONNECT_PASSWORD=supersecret kconnect to uat-bu1
`
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	toCmd := &cobra.Command{
		Use:   "to [historyid/alias/-/LAST/LAST~N]",
		Short: "Re-connect to a previously connected cluster using your history",
		Long: `use is for re-connecting to a previously connected cluster using your history.
You can use the history id or alias as the argument.
You can also supply - or LAST to connect to last cluster in history (current cluster), or LAST~N for previous clusters`,
		Example: examples,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// if len(args) < 1 {
			// 	return ErrAliasIDRequired
			// }

			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Info("running `to` command")

			aliasOrIDORPosition := ""
			if len(args) > 0 {
				aliasOrIDORPosition = args[0]
			}
			params := &app.ConnectToParams{
				Context:             provider.NewContext(provider.WithConfig(cfg)),
				AliasOrIDORPosition: aliasOrIDORPosition,
			}

			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}

			historyLoader, err := loader.NewFileLoader(params.Location)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.Location, err)
			}
			// using to command should never increase number of history items, so set to arbitrary large number
			params.MaxItems = 10000
			store, err := history.NewStore(params.MaxItems, historyLoader)
			if err != nil {
				return fmt.Errorf("creating history store: %w", err)
			}

			a := app.New(app.WithHistoryStore(store))

			return a.ConnectTo(params)
		},
	}

	if err := addConfig(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(toCmd, cfg); err != nil {
		return nil, err
	}

	return toCmd, nil
}

func addConfig(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}
	if _, err := cs.Bool("set-current", true, "Sets the current context in the kubeconfig to the selected cluster"); err != nil {
		return fmt.Errorf("adding set-current config: %w", err)
	}
	if _, err := cs.String("password", "", "Password to use"); err != nil {
		return fmt.Errorf("adding password config: %w", err)
	}
	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location items: %w", err)
	}
	if err := app.AddKubeconfigConfigItems(cs); err != nil {
		return fmt.Errorf("adding kubeconfig config items: %w", err)
	}

	cs.SetHistoryIgnore("password") //nolint
	cs.SetSensitive("password")     //nolint

	return nil
}
