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

	log "github.com/sirupsen/logrus"

	"github.com/fidelity/kconnect/internal/flags"
	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/fidelity/kconnect/pkg/providers/registry"
)

var (
	errMissingProvider    = errors.New("required provider name argument")
	errMissingIdpProtocol = errors.New("missing idp protocol, please use --idp-protocol")
)

type useCmdParams struct {
	Username         string
	Password         string
	Kubeconfig       string
	IdpProtocol      string
	IdpEndpoint      string
	Provider         providers.ClusterProvider
	IdentityProvider providers.IdentityProvider
}

// Command creates the use command
func Command() *cobra.Command {
	params := &useCmdParams{}

	useCmd := &cobra.Command{
		Use:   "use",
		Short: "connect to a target environment and use clusters",
		Args:  cobra.ExactArgs(1),
		AdditionalSetupE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errMissingProvider
			}

			factory := registry.NewProviderFactory()
			selectedProvider, err := factory.CreateClusterProvider(args[0])
			if err != nil {
				return fmt.Errorf("creating cluster provider %s: %w", args[0], err)
			}
			params.Provider = selectedProvider

			log.Infof("using cluster provider %s", params.Provider.Name())
			cmd.Flags().AddFlagSet(params.Provider.Flags())

			idpProtocol := fundIdpProtocolFromArgs(args)
			if idpProtocol == "" {
				return errMissingIdpProtocol
			}
			idProvider, err := factory.CreateIdentityProvider(idpProtocol)
			if err != nil {
				return fmt.Errorf("creating identity provider %s: %w", idpProtocol, err)
			}
			params.IdentityProvider = idProvider
			log.Infof("using identity provider %s", idProvider.Name())
			cmd.Flags().AddFlagSet(params.IdentityProvider.Flags())

			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			identity, err := params.IdentityProvider.Authenticate()
			if err != nil {
				return fmt.Errorf("authenticating using provider %s: %w", params.IdentityProvider.Name(), err)
			}

			return params.Provider.FlagsResolver().Resolve(identity, cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return doUse(cmd, params)
		},
	}

	useCmd.Flags().StringVar(&params.IdpProtocol, "idp-protocol", "", "the idp protocol to use (e.g. saml)")
	flags.AddKubeconfigFlag(useCmd, &params.Kubeconfig)

	useCmd.SetUsageFunc(usage)

	return useCmd
}

func usage(cmd *cobra.Command) error {
	//usage := cmd.UseLine()
	usage := []string{fmt.Sprintf("Usage: %s %s [provider] [flags]", cmd.Parent().CommandPath(), cmd.Use)}

	factory := registry.NewProviderFactory()
	providers := factory.ListClusterProviders()
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

	fmt.Println(strings.Join(usage, "\n"))

	return nil
}

func doUse(c *cobra.Command, params *useCmdParams) error {
	log.WithField("command", "use")

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

func fundIdpProtocolFromArgs(args []string) string {
	index := -1
	for i, arg := range args {
		if arg == "--idp-protocol" {
			index = i
			break
		}
	}
	if index == -1 {
		return ""
	}

	return args[index+1]
}
