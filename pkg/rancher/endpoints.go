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

package rancher

import (
	"fmt"
	"strings"
)

const (
	adAuthTemplate   = "%s-public/activeDirectoryProviders/activedirectory?action=login"
	clustersTemplate = "%s/clusters"
	clusterTemplate  = "%s/clusters/%s"
)

type EndpointsResolver interface {
	ActiveDirectoryAuth() string
	ClustersList() string
	Cluster(clusterName string) string
}

func NewStaticEndpointsResolver(apiEndpoint string) (EndpointsResolver, error) {
	if apiEndpoint == "" {
		return nil, ErrNoAPIEndpoint
	}

	if !strings.HasSuffix(apiEndpoint, "/") {
		apiEndpoint = strings.TrimSuffix(apiEndpoint, "/")
	}

	return &StaticEndpointsResolver{
		apiEndpoint: apiEndpoint,
	}, nil
}

type StaticEndpointsResolver struct {
	apiEndpoint string
}

func (r *StaticEndpointsResolver) ActiveDirectoryAuth() string {
	return fmt.Sprintf(adAuthTemplate, r.apiEndpoint)
}

func (r *StaticEndpointsResolver) ClustersList() string {
	return fmt.Sprintf(clustersTemplate, r.apiEndpoint)
}

func (r *StaticEndpointsResolver) Cluster(clusterName string) string {
	return fmt.Sprintf(clusterTemplate, r.apiEndpoint, clusterName)
}
