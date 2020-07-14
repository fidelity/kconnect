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

package use

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fidelity/kconnect/internal/flags"
	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/fidelity/kconnect/pkg/providers/discovery/aws"
	"github.com/fidelity/kconnect/pkg/providers/identity"
)

var (
	//TODO: this is only temp. There needs to be a reistry
	registeredProviders map[string]providers.ClusterProvider

	errMissingProvider = errors.New("required provider name argument")
	errInvalidProvider = errors.New("invalid provider")
)

type useCmdParams struct {
	Username   string
	Password   string
	Kubeconfig string
	Provider   providers.ClusterProvider
}

// Command creates the use command
func Command() *cobra.Command {
	params := &useCmdParams{}

	useCmd := &cobra.Command{
		Use:   "use",
		Short: "connect to a target environment and use clusters",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errMissingProvider
			}
			selectedProvider := registeredProviders[args[0]]
			if selectedProvider == nil {
				return fmt.Errorf("invalid provider %s: %w", args[0], errInvalidProvider)
			}
			params.Provider = selectedProvider

			fmt.Printf("using provider %s\n", params.Provider.Name())
			cmd.Flags().AddFlagSet(params.Provider.Flags())

			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			fakeIdentity := identity.EmptyIdentityProvider{}

			return params.Provider.FlagsResolver().Resolve(fakeIdentity, cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return doUse(cmd, params)
		},
	}

	flags.AddCommonIdentityFlags(useCmd, &params.Username, &params.Password)
	flags.AddKubeconfigFlag(useCmd, &params.Kubeconfig)

	useCmd.SetUsageFunc(usage)

	return useCmd
}

func init() {
	//TODO: get from provider factory
	registeredProviders = make(map[string]providers.ClusterProvider)
	registeredProviders["eks"] = &aws.EKSClusterProvider{}
}

func usage(cmd *cobra.Command) error {
	//usage := cmd.UseLine()
	usage := []string{fmt.Sprintf("Usage: %s %s [provider] [flags]", cmd.Parent().CommandPath(), cmd.Use)}

	usage = append(usage, "\nProviders:")
	for _, provider := range registeredProviders {
		line := fmt.Sprintf("      %s - %s", provider.Name(), provider.Usage())
		usage = append(usage, strings.TrimRightFunc(line, unicode.IsSpace))
	}

	for _, provider := range registeredProviders {
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

	fmt.Println(strings.Join(usage, "\n"))

	return nil
}

func doUse(c *cobra.Command, params *useCmdParams) error {
	provider := params.Provider

	err := provider.Discover()
	if err != nil {
		return fmt.Errorf("discovering clusters using %s: %w", provider.Name(), err)
	}

	c.Flags().VisitAll(func(flag *pflag.Flag) {
		fmt.Printf("flag %s = %s\n", flag.Name, flag.Value.String())
	})

	return nil
}
