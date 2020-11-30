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
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/github"
)

// Release contains details of a kconnect release
type Release struct {
	Version *string
	Date    *time.Time
	URL     *string
}

// GetLatestRelease gets the latest release detsils from GitHub
func GetLatestRelease() (*Release, error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(context.TODO(), "fidelity", "kconnect")
	if err != nil {
		return nil, fmt.Errorf("getting latest release from GitHub: %w", err)
	}

	return &Release{
		Version: release.TagName,
		Date:    &release.PublishedAt.Time,
		URL:     release.HTMLURL,
	}, nil
}
