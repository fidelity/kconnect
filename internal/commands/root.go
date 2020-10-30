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
	"github.com/fidelity/kconnect/internal/commands/alias"
	"github.com/fidelity/kconnect/internal/commands/configure"
	"github.com/fidelity/kconnect/internal/commands/ls"
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

const (
	shortDesc = "The Kubernetes Connection Manager CLI"
	longDesc  = `The kconnect tool uses a pre-configured Identity Provider to log in to one or
more managed Kubernetes cluster providers, discovers the list of clusters 
visible to your authenticated user and options, and generates a kubectl 
configutation context for the selected cluster.

Most kubectl contexts include an authentication token which kubectl sends to 
Kubernetes with each request rather than a username and password to establish 
your identity.  Authentication tokens typically expire after some time.  The 
user must then to log in again to the managed Kubernetes service provider and 
regenerate the kubectl context for that cluster connection in order to refresh 
the access token.

The kconnect tool makes this much easier by automating the login and kubectl 
context regeneration process, and by allowing the user to repeat previously 
successful connections.

Each time kconnect creates a new connection context, the kconnect tool saves the
information for that connection in the user's connection history list.  The user
can then display their connection history entries and reconnect to any entry by 
its unique ID (or by a user-friendly alias) to refresh an expired access token 
for that cluster.
`
	examples = `# Display a help screen with kconnect commands.
kconnect help

# Configure the default identity provider and connection profile for your user.
#
# Use this command to set up kconnect the first time you use it on a new system.
#
kconnect configure -f FILE_OR_URL

# Create a kubectl confirguration context for an AWS EKS cluster.
#
# Use this command the first time you connect to a new cluster or a new context.
#
kconnect use eks

# Display connection history entries.
#
kconnect ls

# Add an alias to a connection history entry.
#
kconnect alias add --id 012EX456834AFXR0F2NZT68RPKD --alias MYALIAS

# Reconnect and refresh the token for an aliased connection history entry.
#
# Use this to reconnect to a provider and refresh an expired access token.
#
kconnect to MYALIAS

# Display connection history entry aliases.
#
kconnect alias ls
`
)

// RootCmd creates the root kconnect command
func RootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     "kconnect",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
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

	aliasCmd, err := alias.Command()
	if err != nil {
		return nil, fmt.Errorf("creating alias command: %w", err)
	}
	rootCmd.AddCommand(aliasCmd)

	cobra.OnInitialize(initConfig)

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
