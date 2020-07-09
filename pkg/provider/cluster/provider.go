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
package cluster

import (
	"github.com/fidelity/kconnect/pkg/provider/resolver"
	"github.com/urfave/cli/v2"
)

// ClusterProvider is the interface that is used to implement providers
// of Kubernetes clusters. There will be implementations for AWS
// Azure and Rancher initially. Its the job of the provider to
// discover clusters based on a users identity and to generate
// a kubeconfig (and any other files) that are required to access
// a selected cluster.
type ClusterProvider interface {
	// Name is the name of the provider
	Name() string

	// Flags returns the list of CLI flags that are specific to the
	// the provider
	Flags() []cli.Flag

	FlagsResolver() resolver.FlagsResolver

	//Discover(identity identity.Identity, options map[string]string) (map[credentials][]clusters, error)
}

// Cluster represents the information about a discovered k8s cluster
// NOTE: fields in this struct are only for illustration and more though needs to
// go into it
type Cluster interface {
	Name() string

	ID() string

	Region() string
}
