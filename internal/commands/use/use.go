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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/fidelity/kconnect/internal/app"
	"github.com/fidelity/kconnect/internal/usage"
	"github.com/fidelity/kconnect/pkg/provider"
)

var (
	errMissingProvider    = errors.New("required provider name argument")
	errMissingIdpProtocol = errors.New("missing idp protocol, please use --idp-protocol")
)

// type useCmdParams struct {
// 	Kubeconfig       string
// 	IdpProtocol      string
// 	Provider         provider.ClusterProvider
// 	IdentityProvider provider.IdentityProvider
// 	Identity         provider.Identity
// 	Context          *provider.Context
// }

// Command creates the use command
func Command() *cobra.Command {
	logger := logrus.WithField("command", "use")
	a := app.New()

	params := &app.UseParams{}

	useCmd := &cobra.Command{
		Use:   "use",
		Short: "connect to a target environment and use clusters",
		Args:  cobra.ExactArgs(1),
		AdditionalSetupE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errMissingProvider
			}
			return cmdSetup(cmd, args, params, logger)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(cmd, params, logger)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.Use(params)
		},
	}

	useCmd.Flags().StringVar(&params.IdpProtocol, "idp-protocol", "", "the idp protocol to use (e.g. saml)")
	provider.AddKubeconfigFlag(useCmd)
	provider.AddCommonIdentityFlags(useCmd)
	provider.AddCommonClusterProviderFlags(useCmd)

	useCmd.SetUsageFunc(usage.Get)

	return useCmd
}

func cmdSetup(cmd *cobra.Command, args []string, params *app.UseParams, logger *logrus.Entry) error {
	selectedProvider, err := provider.GetClusterProvider(args[0])
	if err != nil {
		return fmt.Errorf("creating cluster provider %s: %w", args[0], err)
	}
	params.Provider = selectedProvider

	params.Context = provider.NewContext(
		cmd,
		provider.WithLogger(logger),
	)

	logger.Infof("using cluster provider %s", params.Provider.Name())
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
	logger.Infof("using identity provider %s", idProvider.Name())
	cmd.Flags().AddFlagSet(params.IdentityProvider.Flags())

	return nil
}

func preRun(cmd *cobra.Command, params *app.UseParams, logger *logrus.Entry) error {
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
