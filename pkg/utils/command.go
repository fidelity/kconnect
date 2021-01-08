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

package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func FormatCommand(cmd *cobra.Command) {

	rootCmdName := "kconnect"
	// If running as a krew plugin, need to change usage output
	if isKrewPlugin() {
		rootCmdName = "kubectl connect"
		// Only change this for root command
		if cmd.Use == "kconnect" {
			cmd.Use = "connect"
		}
		cmd.SetUsageTemplate(strings.NewReplacer(
			"{{.UseLine}}", "kubectl {{.UseLine}}",
			"{{.CommandPath}}", "kubectl {{.CommandPath}}").Replace(cmd.UsageTemplate()))
	}
	cmd.Example = formatMessage(cmd.Example, rootCmdName)
}

func isKrewPlugin() bool {
	return strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-")
}

func formatMessage(message, rootCmdName string) string {
	return strings.NewReplacer("{{.CommandPath}}", rootCmdName).Replace(message)
}
