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

package provider

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/client-go/tools/clientcmd"
)

// ClusterProviderConfig represents the base configuration for
// a cluster provider
type ClusterProviderConfig struct {
	ClusterName *string `flag:"cluster"`
}

// IdentityProviderConfig represents the base configuration for an
// identity provider.
type IdentityProviderConfig struct {
	Username string `flag:"username" validate:"required"`
	Password string `flag:"password" validate:"required"`
	Force    bool   `flag:"force"`
}

// AddCommonClusterProviderFlags will add flags that are common to
// all cluster providers
func AddCommonClusterProviderFlags(cmd *cobra.Command) {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringP("cluster", "c", "", "Name of the cluster to use.")

	cmd.Flags().AddFlagSet(fs)
}

// AddCommonIdentityFlags will add common identity related flags to a command
func AddCommonIdentityFlags(cmd *cobra.Command) {
	flagSet := CommonIdentityFlagSet()
	cmd.Flags().AddFlagSet(flagSet)
}

// CommonIdentityFlagSet creates a flagset with the common identity flags
func CommonIdentityFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("username", "", "the username used for authentication")
	flagSet.String("password", "", "the password to use for authentication")
	flagSet.Bool("force", false, "If true then we force authentication every invocation")

	return flagSet
}

// AddKubeconfigFlag will add the kubeconfig flag to a command
func AddKubeconfigFlag(cmd *cobra.Command) {
	flagSet := &pflag.FlagSet{}

	pathOptions := clientcmd.NewDefaultPathOptions()
	flagSet.StringP("kubeconfig", "k", pathOptions.GetDefaultFilename(), "location of the kubeconfig to use")

	cmd.Flags().AddFlagSet(flagSet)
}
