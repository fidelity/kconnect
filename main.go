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

package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fidelity/kconnect/internal/commands/version"
)

var configFile string

func main() {

	rootCmd := cobra.Command{
		Use:   "kconnect [command]",
		Short: "The Kubernetes Connection Manager CLI",
		Run: func(c *cobra.Command, _ []string) {
			if err := c.Help(); err != nil {
				log.Debugf("ignoring error %s", err.Error())
			}
		},
	}

	rootCmd.PersistentFlags().BoolP("help", "h", false, "help for a command")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "", "", "Configuration file (defaults to $HOME/.kconnect/config)")

	rootCmd.AddCommand(version.Command())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
