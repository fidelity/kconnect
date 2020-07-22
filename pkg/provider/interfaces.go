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
	"github.com/spf13/pflag"
)

// ClusterProvider is the interface that is used to implement providers
// of Kubernetes clusters. There will be implementations for AWS
// Azure and Rancher initially. Its the job of the provider to
// discover clusters based on a users identity and to generate
// a kubeconfig (and any other files) that are required to access
// a selected cluster.
type ClusterProvider interface {
	Plugin

	FlagsResolver() FlagsResolver

	// NOTE: this needs to return something like map[credentials][]clusters
	// and also have identity and options passed in
	Discover(ctx *Context, identity Identity) error
}

// FlagsResolver is used to resolve the values for flags interactively.
// There will be a flags resolver for Azure, AWS and Rancher initially.
type FlagsResolver interface {

	// Resolve will resolve the values for the supplied context. It will interactively
	// resolve the values by asking the user for selections. It will basically set the
	// Value on each pflag.Flag
	Resolve(ctx *Context, identity Identity) error
}

// IdentityProvider represents the interface used to implement an identity provider
// plugin. It provides authentication and authorization functionality.
type IdentityProvider interface {
	Plugin

	// Authenticate will authenticate a user and return details of
	// their identity.
	Authenticate(ctx *Context) (Identity, error)
}

// Cluster represents the information about a discovered k8s cluster
// NOTE: fields in this struct are only for illustration and more though needs to
// go into it
type Cluster interface {
	Name() string

	ID() string

	Region() string
}

// Identity represents a users identity for use with discovery.
// NOTE: details of this need finalising
type Identity interface {
}

// Plugin is an interface that can be implemented for returned usage information
type Plugin interface {
	// Name returns the name of the plugin
	Name() string

	// Flags will return the flags as part of a flagset for the plugin
	Flags() *pflag.FlagSet

	// Usage returns a string to display for help
	Usage() string
}
