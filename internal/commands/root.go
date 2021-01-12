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
	"time"

	"github.com/blang/semver"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/internal/commands/alias"
	configcmd "github.com/fidelity/kconnect/internal/commands/config"
	"github.com/fidelity/kconnect/internal/commands/logout"
	"github.com/fidelity/kconnect/internal/commands/ls"
	"github.com/fidelity/kconnect/internal/commands/to"
	"github.com/fidelity/kconnect/internal/commands/use"
	"github.com/fidelity/kconnect/internal/commands/version"
	"github.com/fidelity/kconnect/internal/defaults"
	appver "github.com/fidelity/kconnect/internal/version"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/utils"
)

var (
	cfg                  config.ConfigurationSet
	versionCheckInterval time.Duration = 1440 * time.Minute
)

const (
	shortDesc = "The Kubernetes Connection Manager CLI"
	longDesc  = `
The kconnect tool uses a pre-configured Identity Provider to log in to one or
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
	examplesTemplate = `
  # Display a help screen with kconnect commands.
  {{.CommandPath}} help

  # Configure the default identity provider and connection profile for your user.
  #
  # Use this command to set up kconnect the first time you use it on a new system.
  #
  {{.CommandPath}} config -f FILE_OR_URL

  # Create a kubectl confirguration context for an AWS EKS cluster.
  #
  # Use this command the first time you connect to a new cluster or a new context.
  #
  {{.CommandPath}} use eks

  # Display connection history entries.
  #
  {{.CommandPath}} ls

  # Add an alias to a connection history entry.
  #
  {{.CommandPath}} alias add --id 012EX456834AFXR0F2NZT68RPKD --alias MYALIAS

  # Reconnect and refresh the token for an aliased connection history entry.
  #
  # Use this to reconnect to a provider and refresh an expired access token.
  #
  {{.CommandPath}} to MYALIAS

  # Display connection history entry aliases.
  #
  {{.CommandPath}} alias ls
`
)

// RootCmd creates the root kconnect command
func RootCmd() (*cobra.Command, error) {
	cfg = config.NewConfigurationSet()

	rootCmd := &cobra.Command{
		Use:     "kconnect",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examplesTemplate,
		Run: func(c *cobra.Command, _ []string) {
			if err := c.Help(); err != nil {
				zap.S().Debugw("ingoring cobra error",
					"error",
					err.Error())
			}
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			inTerminal := isRunningInTerminal()
			if !inTerminal {
				zap.S().Debug("Not running in a terminal, setting no-input to true")
				cmd.Flags().Set(app.NoInputConfigItem, "true") //nolint: errcheck
				return nil
			}

			if err := flags.CopyFlagValue(app.NonInteractiveConfigItem, app.NoInputConfigItem, cmd.Flags(), true); err != nil {
				return fmt.Errorf("copying flag value from %s to %s: %w", app.NonInteractiveConfigItem, app.NoInputConfigItem, err)
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			params := &app.CommonConfig{}
			if err := config.Unmarshall(cfg, params); err != nil {
				return fmt.Errorf("unmarshalling config into to params: %w", err)
			}
			if !params.DisableVersionCheck {
				if err := reportNewerVersion(); err != nil {
					return fmt.Errorf("reporting newer version: %w", err)
				}
			}

			return nil
		},
	}
	utils.FormatCommand(rootCmd)

	if err := ensureAppDirectory(); err != nil {
		return nil, fmt.Errorf("ensuring app directory exists: %w", err)
	}

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
	cfgCmd, err := configcmd.Command()
	if err != nil {
		return nil, fmt.Errorf("creating config command: %w", err)
	}
	rootCmd.AddCommand(cfgCmd)
	rootCmd.AddCommand(version.Command())

	aliasCmd, err := alias.Command()
	if err != nil {
		return nil, fmt.Errorf("creating alias command: %w", err)
	}
	rootCmd.AddCommand(aliasCmd)

	logoutCmd, err := logout.Command()
	if err != nil {
		return nil, fmt.Errorf("creating logout command: %w", err)
	}
	rootCmd.AddCommand(logoutCmd)

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

func isRunningInTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}

func reportNewerVersion() error {
	appCfg, err := config.NewAppConfiguration()
	if err != nil {
		return fmt.Errorf("creating app configuration: %w", err)
	}

	cfg, err := appCfg.Get()
	if err != nil {
		return fmt.Errorf("getting configuration: %w", err)
	}

	v := appver.Get()
	if v.Version == "" {
		// Running a local build so set the version number
		v.Version = "0.0.0"
	}

	currentSemver, err := semver.Parse(v.Version)
	if err != nil {
		return fmt.Errorf("parsing current version %s: %w", v.Version, err)
	}

	var latestSemver semver.Version
	checkTime := time.Now().UTC()
	checkDiff := checkTime.Sub(cfg.Spec.VersionCheck.LastChecked.Time)
	if checkDiff > versionCheckInterval { //nolint:nestif
		latestRelease, err := appver.GetLatestRelease()
		if err != nil {
			return fmt.Errorf("getting latest release: %w", err)
		}

		latestSemver, err = semver.Parse(*latestRelease.Version)
		if err != nil {
			return fmt.Errorf("parsing latest release version %s: %w", *latestRelease.Version, err)
		}

		if latestSemver.GT(currentSemver) {
			cfg.Spec.VersionCheck.LatestReleaseVersion = latestRelease.Version
			cfg.Spec.VersionCheck.LatestReleaseURL = latestRelease.URL
		}

		cfg.Spec.VersionCheck.LastChecked = metav1.NewTime(checkTime)
		if err := appCfg.Save(cfg); err != nil {
			return fmt.Errorf("saving app configuration: %w", err)
		}
	} else {
		zap.S().Debugw("latest version not retrieved as check interval not exceeded", "diffMins", checkDiff.Minutes(), "savedVersion", cfg.Spec.VersionCheck.LatestReleaseVersion)
		if cfg.Spec.VersionCheck.LatestReleaseVersion != nil && *cfg.Spec.VersionCheck.LatestReleaseVersion != "" {
			latestSemver, err = semver.Parse(*cfg.Spec.VersionCheck.LatestReleaseVersion)
			if err != nil {
				return fmt.Errorf("parsing saved latest release version %s: %w", *cfg.Spec.VersionCheck.LatestReleaseVersion, err)
			}
		}
	}

	if latestSemver.GT(currentSemver) {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintf(os.Stderr, "\033[33mNew kconnect version available: v%s -> v%s\033[0m\n", currentSemver.String(), latestSemver.String())
		fmt.Fprintf(os.Stderr, "\033[33mVisit %s for more details\033[0m\n", *cfg.Spec.VersionCheck.LatestReleaseURL)
	}

	return nil
}
