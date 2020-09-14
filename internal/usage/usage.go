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

package usage

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/spf13/cobra"
)

// Get is used to print the usage for a command
func Get(cmd *cobra.Command) error {
	usage := []string{fmt.Sprintf("Usage: %s %s [provider] [flags]", cmd.Parent().CommandPath(), cmd.Use)}

	providers := provider.ListClusterProviders()
	usage = append(usage, "\nProviders:")
	for _, provider := range providers {
		line := fmt.Sprintf("      %s - %s", provider.Name(), provider.Usage())
		usage = append(usage, strings.TrimRightFunc(line, unicode.IsSpace))
	}

	for _, provider := range providers {
		if provider.ConfigurationItems() != nil {
			usage = append(usage, fmt.Sprintf("\n%s provider flags:", provider.Name()))
			providerFlags, err := flags.CreateFlagsFromConfig(provider.ConfigurationItems())
			if err != nil {
				return fmt.Errorf("converting provider config to flags: %w", err)
			}
			usage = append(usage, strings.TrimRightFunc(providerFlags.FlagUsages(), unicode.IsSpace))
		}
	}

	usage = append(usage, "\nCommon Flags:")
	if len(cmd.PersistentFlags().FlagUsages()) != 0 {
		usage = append(usage, strings.TrimRightFunc(cmd.PersistentFlags().FlagUsages(), unicode.IsSpace))
	}
	if len(cmd.InheritedFlags().FlagUsages()) != 0 {
		usage = append(usage, strings.TrimRightFunc(cmd.InheritedFlags().FlagUsages(), unicode.IsSpace))
	}

	fmt.Printf("%s\n", strings.Join(usage, "\n"))

	return nil
}
