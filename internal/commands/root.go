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

package commands

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/internal/commands/configure"
	"github.com/fidelity/kconnect/internal/commands/ls"
	"github.com/fidelity/kconnect/internal/commands/renew"
	"github.com/fidelity/kconnect/internal/commands/to"
	"github.com/fidelity/kconnect/internal/commands/use"
	"github.com/fidelity/kconnect/internal/commands/version"
	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
)

var (
	cfg config.ConfigurationSet
)

// RootCmd creates the root kconnect command
func RootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "kconnect",
		Short: "The Kubernetes Connection Manager CLI",
		Run: func(c *cobra.Command, _ []string) {
			if err := c.Help(); err != nil {
				zap.S().Debugw("ingoring cobra error",
					"error",
					err.Error())
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	if err := ensureAppDirectory(); err != nil {
		return nil, fmt.Errorf("ensuring app directory exists: %w", err)
	}

	cfg = config.NewConfigurationSet()
	if err := app.AddCommonConfigItems(cfg); err != nil {
		return nil, fmt.Errorf("adding common configuration: %w", err)
	}
	rootFlags, err := flags.CreateFlagsFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating root command flags: %w", err)
	}
	rootCmd.PersistentFlags().AddFlagSet(rootFlags)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	useCmd, err := use.Command()
	if err != nil {
		return nil, fmt.Errorf("creating use command: %w", err)
	}
	rootCmd.AddCommand(useCmd)
	toCmd, err := to.Command()
	if err != nil {
		return nil, fmt.Errorf("creating to command: %w", err)
	}
	rootCmd.AddCommand(toCmd)
	lsCmd, err := ls.Command()
	if err != nil {
		return nil, fmt.Errorf("creating ls command: %w", err)
	}
	rootCmd.AddCommand(lsCmd)
	cfgCmd, err := configure.Command()
	if err != nil {
		return nil, fmt.Errorf("creating configure command: %w", err)
	}
	rootCmd.AddCommand(cfgCmd)
	rootCmd.AddCommand(version.Command())

	renewCmd, err := renew.Command()
	if err != nil {
		return nil, fmt.Errorf("creating renew command: %w", err)
	}
	rootCmd.AddCommand(renewCmd)

	cobra.OnInitialize(initConfig)

	// Force initial parsing of flags
	rootCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{
		UnknownFlags: true,
	}
	rootCmd.ParseFlags(os.Args) //nolint: errcheck

	flags.PopulateConfigFromCommand(rootCmd, cfg)
	params := &app.CommonConfig{}
	if err := config.Unmarshall(cfg, params); err != nil {
		return nil, fmt.Errorf("unmarshalling config into use params: %w", err)
	}

	return rootCmd, nil
}

func initConfig() {
	viper.SetEnvPrefix("KCONNECT")
	viper.AutomaticEnv()
}

func ensureAppDirectory() error {
	appDir := defaults.AppDirectory()

	_, err := os.Stat(appDir)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("getting details of app directory %s: %w", appDir, err)
	}

	if err := os.Mkdir(appDir, os.ModePerm); err != nil {
		return fmt.Errorf("making app folder directory %s: %w", appDir, err)
	}

	return nil
}
