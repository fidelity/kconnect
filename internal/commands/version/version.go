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

package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/fidelity/kconnect/internal/version"
)

// Command creates the version cobra command
func Command() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version & build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doVersion(cmd)
		},
	}

	return versionCmd
}

func doVersion(_ *cobra.Command) error {
	v := version.Get()
	outYaml, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshalling version information: %w", err)
	}
	fmt.Println(string(outYaml)) //nolint:forbidigo

	return nil
}
