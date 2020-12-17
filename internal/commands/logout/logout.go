package logout

import (
	"fmt"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	shortDesc = "Query the user's connection history"
	longDesc  = `
Query and display the user's connection history entries, including entry IDs and
aliases.

Each time kconnect creates a new kubectl context to connect to a Kubernetes 
cluster, it saves the settings for the new connection as an entry in the user's 
connection history.  The user can then reconnect using those same settings later 
via the connection history entry's ID or alias.
`
)

func Command() (*cobra.Command, error) {
	cfg := config.NewConfigurationSet()

	logoutCmd := &cobra.Command{
		Use:     "logout",
		Short:   shortDesc,
		Long:    longDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flags.BindFlags(cmd)
			flags.PopulateConfigFromCommand(cmd, cfg)

			if err := config.ApplyToConfigSet(cfg); err != nil {
				return fmt.Errorf("applying app config: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zap.S().Debug("running `ls` command")
			params := &app.LogoutParams{
				Context: provider.NewContext(provider.WithConfig(cfg)),
			}

			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into logout params: %w", err)
			}

			historyLoader, err := loader.NewFileLoader(params.Location)
			if err != nil {
				return fmt.Errorf("getting history loader with path %s: %w", params.Location, err)
			}
			store, err := history.NewStore(params.MaxItems, historyLoader)
			if err != nil {
				return fmt.Errorf("creating history store: %w", err)
			}

			a := app.New(app.WithHistoryStore(store))

			return a.Logout(params)
		},
	}

	if err := addConfig(cfg); err != nil {
		return nil, fmt.Errorf("add command config: %w", err)
	}

	if err := flags.CreateCommandFlags(logoutCmd, cfg); err != nil {
		return nil, err
	}

	return logoutCmd, nil

}


func addConfig(cs config.ConfigurationSet) error {
	if err := app.AddCommonConfigItems(cs); err != nil {
		return fmt.Errorf("adding common config: %w", err)
	}

	if _, err := cs.Bool("all", false, "Logs out of all clusters"); err != nil {
		return fmt.Errorf("adding all config: %w", err)
	}
	if err := cs.SetShort("all", "a"); err != nil {
		return fmt.Errorf("adding all short flag: %w", err)
	}
	if _, err := cs.String("alias", "", "comma delimited list of aliass"); err != nil {
		return fmt.Errorf("adding alias config: %w", err)
	}
	if _, err := cs.String("ids", "", "comma delimited list of ids"); err != nil {
		return fmt.Errorf("adding ids config: %w", err)
	}

	if err := app.AddHistoryLocationItems(cs); err != nil {
		return fmt.Errorf("adding history location items: %w", err)
	}
	if err := app.AddKubeconfigConfigItems(cs); err != nil {
		return fmt.Errorf("adding kubeconfig config items: %w", err)
	}

	return nil
}