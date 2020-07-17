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

package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/client-go/tools/clientcmd"
)

// AddCommonIdentityFlags will add common identity related flags to a command
func AddCommonIdentityFlags(cmd *cobra.Command, username, password *string) {
	flagSet := CommonIdentityFlagSet(username, password)
	cmd.Flags().AddFlagSet(flagSet)
}

// CommonIdentityFlagSet creates a flagset with the common identity flags
func CommonIdentityFlagSet(username, password *string) *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVar(username, "username", "", "the username used for authentication")
	flagSet.StringVar(password, "password", "", "the password to use for authentication")

	return flagSet
}

// AddKubeconfigFlag will add the kubeconfig flag to a command
func AddKubeconfigFlag(cmd *cobra.Command, kubeconfig *string) {
	flagSet := &pflag.FlagSet{}

	pathOptions := clientcmd.NewDefaultPathOptions()
	flagSet.StringP("kubeconfig", "k", pathOptions.GetDefaultFilename(), "location of the kubeconfig to use")

	cmd.Flags().AddFlagSet(flagSet)
}
