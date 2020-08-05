package usage

import (
	"fmt"
	"strings"
	"unicode"

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
		if provider.Flags() != nil {
			usage = append(usage, fmt.Sprintf("\n%s provider flags:", provider.Name()))
			usage = append(usage, strings.TrimRightFunc(provider.Flags().FlagUsages(), unicode.IsSpace))
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
