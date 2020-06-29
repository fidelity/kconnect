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
	"runtime"

	"github.com/spf13/cobra"

	"github.com/fidelity/kconnect/internal/version"
)

// Command creates the version command
func Command() *cobra.Command {

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version & build information",
		Long:  "",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return doVersion(cmd)
		},
	}

	return versionCmd
}

func doVersion(_ *cobra.Command) error {

	fmt.Printf("Version: %s\n", version.Version)
	fmt.Printf("Commit: %s\n", version.CommitHash)
	fmt.Printf("Date: %s\n", version.BuildDate)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("GOOS: %s\n", runtime.GOOS)
	fmt.Printf("GOARCH: %s\n", runtime.GOARCH)

	return nil
}
