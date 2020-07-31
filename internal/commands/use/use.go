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
	"strconv"
	"strings"
	"unicode"

	survey "github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"github.com/fidelity/kconnect/pkg/provider"
)

const (
	defaultPageSize = 10
)

var (
	errMissingProvider    = errors.New("required provider name argument")
	errMissingIdpProtocol = errors.New("missing idp protocol, please use --idp-protocol")

	logger = log.WithField("command", "use")
)

type useCmdParams struct {
	Kubeconfig       string
	IdpProtocol      string
	Provider         provider.ClusterProvider
	IdentityProvider provider.IdentityProvider
	Identity         provider.Identity
	Context          *provider.Context
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

			selectedProvider, err := provider.GetClusterProvider(args[0])
			if err != nil {
				return fmt.Errorf("creating cluster provider %s: %w", args[0], err)
			}
			params.Provider = selectedProvider

			params.Context = provider.NewContext(
				cmd,
				provider.WithLogger(logger),
			)

			log.Infof("using cluster provider %s", params.Provider.Name())
			cmd.Flags().AddFlagSet(params.Provider.Flags())

			idpProtocol := findIdpProtocolFromArgs(args)
			if idpProtocol == "" {
				return errMissingIdpProtocol
			}
			idProvider, err := provider.GetIdentityProvider(idpProtocol)
			if err != nil {
				return fmt.Errorf("creating identity provider %s: %w", idpProtocol, err)
			}
			params.IdentityProvider = idProvider
			log.Infof("using identity provider %s", idProvider.Name())
			cmd.Flags().AddFlagSet(params.IdentityProvider.Flags())

			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Update the context now the flags have been parsed
			interactive, err := isInteractive(cmd)
			if err != nil {
				return fmt.Errorf("getting interactive flag: %w", err)
			}
			params.Context = provider.NewContext(
				cmd,
				provider.WithLogger(logger),
				provider.WithInteractive(interactive),
			)

			identity, err := params.IdentityProvider.Authenticate(params.Context, params.Provider.Name())
			if err != nil {
				return fmt.Errorf("authenticating using provider %s: %w", params.IdentityProvider.Name(), err)
			}
			params.Identity = identity

			return params.Provider.FlagsResolver().Resolve(params.Context, cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return doUse(params)
		},
	}

	useCmd.Flags().StringVar(&params.IdpProtocol, "idp-protocol", "", "the idp protocol to use (e.g. saml)")
	provider.AddKubeconfigFlag(useCmd)
	//TODO: should these be added in the actual plugins???
	provider.AddCommonIdentityFlags(useCmd)
	provider.AddCommonClusterProviderFlags(useCmd)

	useCmd.SetUsageFunc(usage)

	return useCmd
}

func usage(cmd *cobra.Command) error {
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

func doUse(params *useCmdParams) error {
	logger.Info("Executing command")

	provider := params.Provider

	discoverOutput, err := provider.Discover(params.Context, params.Identity)
	if err != nil {
		return fmt.Errorf("discovering clusters using %s: %w", provider.Name(), err)
	}

	if discoverOutput.Clusters == nil || len(discoverOutput.Clusters) == 0 {
		logger.Info("no clusters discovered, not continuing")
	}

	cluster, err := selectCluster(discoverOutput)
	if err != nil {
		return fmt.Errorf("selecting cluster: %w", err)
	}

	kubeConfig, err := provider.GetClusterConfig(params.Context, cluster)
	if err != nil {
		return fmt.Errorf("creating kubeconfig for %s: %w", cluster.Name, err)
	}

	if err := kubeconfig.Write(params.Kubeconfig, kubeConfig); err != nil {
		return fmt.Errorf("writing cluster kubeconfig: %w", err)
	}

	return nil
}

func selectCluster(discoverOutput *provider.DiscoverOutput) (*provider.Cluster, error) {
	options := []string{}
	for _, cluster := range discoverOutput.Clusters {
		options = append(options, cluster.Name)
	}

	if len(options) == 1 {
		return discoverOutput.Clusters[options[0]], nil
	}

	clusterName := ""
	prompt := &survey.Select{
		Message:  "Select the cluster",
		Options:  options,
		PageSize: defaultPageSize,
		Help:     "Select a cluster to connect to from the discovered clusters",
	}
	if err := survey.AskOne(prompt, &clusterName, survey.WithValidator(survey.Required)); err != nil {
		return nil, fmt.Errorf("selecting a cluster: %w", err)
	}

	return discoverOutput.Clusters[clusterName], nil
}

func findIdpProtocolFromArgs(args []string) string {
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

func isInteractive(cmd *cobra.Command) (bool, error) {
	flag := cmd.Flags().Lookup("non-interactive")
	if flag == nil {
		return false, nil
	}

	ni, err := strconv.ParseBool(flag.Value.String())
	if err != nil {
		return false, fmt.Errorf("parsing interactive flag: %w", err)
	}

	return !ni, nil
}
