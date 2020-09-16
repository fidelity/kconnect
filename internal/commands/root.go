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

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/internal/commands/ls"
	"github.com/fidelity/kconnect/internal/commands/to"
	"github.com/fidelity/kconnect/internal/commands/use"
	"github.com/fidelity/kconnect/internal/commands/version"
	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
				logrus.Debugf("ignoring cobra error %q", err.Error())
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
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
	rootCmd.AddCommand(version.Command())

	flags.PopulateConfigFromCommand(rootCmd, cfg)
	params := &app.CommonConfig{}
	if err := config.Unmarshall(cfg, params); err != nil {
		return nil, fmt.Errorf("unmarshalling config into use params: %w", err)
	}

	cobra.OnInitialize(initConfig(params))

	// err := logging.Configure(logLevel, logFormat)
	// if err != nil {
	// 	return fmt.Errorf("configuring logging: %w", err)
	// }

	return rootCmd, nil
}

func initConfig(cfg *app.CommonConfig) func() {
	return func() {
		if cfg.ConfigFile != "" {
			viper.SetConfigFile(cfg.ConfigFile)
		} else {
			home := defaults.AppDirectory()

			viper.AddConfigPath(home)
			viper.SetConfigName("config")
		}

		viper.SetEnvPrefix("KCONNECT")
		viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err == nil {
			logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
		}
	}
}
