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

package rancher_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/fidelity/kconnect/pkg/rancher"
)

func TestStaticEndpointsResolverTrailingSlash(t *testing.T) {
	withSlash, err := rancher.NewStaticEndpointsResolver("https://rancher.example.com/v3/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	withoutSlash, err := rancher.NewStaticEndpointsResolver("https://rancher.example.com/v3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	endpoints := map[string][2]string{
		"active directory": {withSlash.ActiveDirectoryAuth(), withoutSlash.ActiveDirectoryAuth()},
		"clusters":         {withSlash.ClustersList(), withoutSlash.ClustersList()},
		"cluster":          {withSlash.Cluster("c-abc"), withoutSlash.Cluster("c-abc")},
	}

	for name, urls := range endpoints {
		if urls[0] != urls[1] {
			t.Errorf("%s: a trailing slash changed the url: %q and %q", name, urls[0], urls[1])
		}

		if strings.Contains(strings.TrimPrefix(urls[0], "https://"), "//") {
			t.Errorf("%s: url has an empty path segment: %q", name, urls[0])
		}

		// The public endpoints are under /v3-public and the others under /v3/
		if !strings.HasPrefix(urls[0], "https://rancher.example.com/v3") {
			t.Errorf("%s: unexpected url %q", name, urls[0])
		}
	}
}

func TestStaticEndpointsResolverNeedsAnEndpoint(t *testing.T) {
	if _, err := rancher.NewStaticEndpointsResolver(""); !errors.Is(err, rancher.ErrNoAPIEndpoint) {
		t.Fatalf("expected ErrNoAPIEndpoint, got %v", err)
	}
}
