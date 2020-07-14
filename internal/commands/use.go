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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/fidelity/kconnect/pkg/providers/discovery/aws"
	"github.com/fidelity/kconnect/pkg/providers/identity"
)

var (
	//TODO: this is only temp. There needs to be a reistry
	registeredProviders map[string]providers.ClusterProvider
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "connect to a target environment and use clusters",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("required provider name argument")
		}
		selectedProvider := registeredProviders[args[0]]
		if selectedProvider == nil {
			return fmt.Errorf("invalid provider: %s", args[0])
		}

		fmt.Printf("using provider %s\n", selectedProvider.Name())
		cmd.Flags().AddFlagSet(selectedProvider.Flags())
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		fakeIdentity := identity.EmptyIdentityProvider{}

		selectedProvider := registeredProviders[args[0]]
		return selectedProvider.FlagsResolver().Resolve(fakeIdentity, cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		selectedProvider := registeredProviders[args[0]]
		return doUse(cmd, selectedProvider)
	},
}

func init() {
	// Add flags that are common across all
	useCmd.Flags().AddFlagSet(commonUseFlags())
	useCmd.MarkFlagRequired("username")

	//useCmd.SetUsageFunc(usage)

	//TODO: get from provider factory
	registeredProviders = make(map[string]providers.ClusterProvider)
	registeredProviders["eks"] = &aws.EKSClusterProvider{}

	RootCmd.AddCommand(useCmd)
}

func usage(cmd *cobra.Command) error {
	usage := cmd.UseLine()
	fmt.Println(usage)

	return nil
}

func commonUseFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("username", "", "the username used for authentication")

	return flagSet
}

func doUse(c *cobra.Command, provider providers.ClusterProvider) error {
	fmt.Println("In do Use\n")
	fmt.Printf("With provider: %s\n", provider.Name())

	c.Flags().VisitAll(func(flag *pflag.Flag) {
		fmt.Printf("flag %s = %s\n", flag.Name, flag.Value.String())
	})

	return nil
}
