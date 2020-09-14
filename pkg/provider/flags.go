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
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/client-go/tools/clientcmd"
)

// ClusterProviderConfig represents the base configuration for
// a cluster provider
type ClusterProviderConfig struct {
	ClusterName *string `flag:"cluster" json:"cluster"`
}

// IdentityProviderConfig represents the base configuration for an
// identity provider.
type IdentityProviderConfig struct {
	Username string `flag:"username" validate:"required" json:"username"`
	Password string `flag:"password" validate:"required" json:"password"`
	Force    bool   `flag:"force" json:"force"`
}

// AddCommonClusterProviderFlags will add flags that are common to
// all cluster providers
func AddCommonClusterProviderFlags(cmd *cobra.Command) {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringP("cluster", "c", "", "Name of the cluster to use.")

	cmd.Flags().AddFlagSet(fs)
}

func AddCommonClusterConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("cluster", "", "Name of the cluster to use."); err != nil {
		return fmt.Errorf("adding cluster setting: %w", err)
	}
	if _, err := cs.Bool("non-interactive", false, "Run without interactive flag resolution. Defaults to false"); err != nil {
		return fmt.Errorf("adding non-interactive setting: %w", err)
	}
	if err := cs.SetShort("cluster", "c"); err != nil {
		return fmt.Errorf("setting shorthand for cluster setting: %w", err)
	}

	return nil
}

// AddCommonIdentityFlags will add common identity related flags to a command
func AddCommonIdentityFlags(cmd *cobra.Command) {
	flagSet := CommonIdentityFlagSet()
	cmd.Flags().AddFlagSet(flagSet)
}

// AddCommonIdentityFlags will add common identity related flags to a command
func AddCommonIdentityConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("username", "", "the username used for authentication"); err != nil {
		return fmt.Errorf("adding username config: %w", err)
	}
	if _, err := cs.String("password", "", "the password to use for authentication"); err != nil {
		return fmt.Errorf("adding password config: %w", err)
	}
	if _, err := cs.Bool("force", false, "If true then we force authentication every invocation"); err != nil {
		return fmt.Errorf("adding force config: %w", err)
	}

	return nil
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

func AddKubeconfigConfig(cs config.ConfigurationSet) error {
	pathOptions := clientcmd.NewDefaultPathOptions()
	if _, err := cs.String("kubeconfig", pathOptions.GetDefaultFilename(), "location of the kubeconfig to use"); err != nil {
		return fmt.Errorf("adding kubeconfig config: %w", err)
	}
	if err := cs.SetShort("kubeconfig", "k"); err != nil {
		return fmt.Errorf("setting shorthand for kubeconfig config: %w", err)
	}

	return nil
}
