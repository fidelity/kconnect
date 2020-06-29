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
)

var (
	version    string // Specifies the app version
	buildDate  string // Specifies the build date
	commitHash string // Specifies the git commit hash
)

// Info represents the version information for the app
type Info struct {
	Version    string `json:"version,omitempty"`
	BuildDate  string `json:"buildDate,omitempty"`
	CommitHash string `json:"commitHash,omitempty"`
	GoVersion  string `json:"goVersion,omitempty"`
	Platform   string `json:"platform,omitempty"`
	Compiler   string `json:"compiler,omitempty"`
}

// String will convert the version information to a string
func (i Info) String() string {
	return fmt.Sprintf("Version: %s, Build Date: %s, Git Hash: %s", i.Version, i.BuildDate, i.CommitHash)
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:    version,
		BuildDate:  buildDate,
		CommitHash: commitHash,
		GoVersion:  runtime.Version(),
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Compiler:   runtime.Compiler,
	}
}
