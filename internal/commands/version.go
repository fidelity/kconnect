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
	"fmt"

	"github.com/fidelity/kconnect/internal/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version & build information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doVersion(cmd)
	},
}

func init() {
	// TODO: add any additional flags
	RootCmd.AddCommand(versionCmd)
}

func doVersion(c *cobra.Command) error {
	v := version.Get()
	outYaml, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Println(string(outYaml))

	return nil
}
