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

	shortDesc = "Reconnect to a connection history entry."
	longDesc  = `
Reconnect to a cluster in the connection history by its entry ID or alias.

The kconnect tool creates an entry in the user's connection history with all the
connection settings each time it generates a new kubectl configuration context
for a Kubernetes cluster.  The user can then reconnect to the same cluster and
refresh their access token or regenerate the kubectl configuration context using
the connection history entry's ID or alias.

The to command also accepts - or LAST as proxy references to the most recent
connection history entry, or LAST~N for the Nth previous entry.

Although kconnect does not save the user's password in the connection history,
the user can avoid having to enter their password interactively by setting the
KCONNECT_PASSWORD environment variable or the --password command-line flag.
Otherwise kconnect will promot the user to enter their password.
`
	examples = `
  # Reconnect based on an alias - aliases can be found using kconnect ls
  kconnect to uat-bu1

  # Reconnect based on an history id - history id can be found using kconnect ls
  kconnect to 01EM615GB2YX3C6WZ9MCWBDWBF

  # Re-connect interactively from history list
  kconnect to

  # Re-connect to current cluster (this is useful for renewing credentials)
  kconnect to -
  OR
  kconnect to LAST

  # Re-connect to cluster used before current one
  kconnect to LAST~1

  # Re-connect based on an alias supplying a password
  kconnect to uat-bu1 --password supersecret

  # Re-connect based on an alias supplying a password via env var
  KCONNECT_PASSWORD=supersecret kconnect to uat-bu2
 `
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	toCmd := &cobra.Command{
		Use:     "to [historyid/alias/-/LAST/LAST~N]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
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
