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
	"k8s.io/client-go/tools/clientcmd/api"
)

// ClusterProvider is the interface that is used to implement providers
// of Kubernetes clusters. There will be implementations for AWS
// Azure and Rancher initially. Its the job of the provider to
// discover clusters based on a users identity and to generate
// a kubeconfig (and any other files) that are required to access
// a selected cluster.
type ClusterProvider interface {
	Plugin

	// Discover will discover what clusters the supplied identity has access to
	Discover(ctx *Context, identity Identity) (*DiscoverOutput, error)

	GetClusterConfig(ctx *Context, cluster *Cluster, setCurrent bool) (*api.Config, error)

	// FlagsResolver returns the resolver used to inetractively resolve flags
	FlagsResolver() FlagsResolver
}

// FlagsResolver is used to resolve the values for flags interactively.
// There will be a flags resolver for Azure, AWS and Rancher initially.
type FlagsResolver interface {

	// Validate is used to validate the flags and return any errors
	Validate(ctx *Context, flags *pflag.FlagSet) error

	// Resolve will resolve the values for the supplied context. It will interactively
	// resolve the values by asking the user for selections. It will basically set the
	// Value on each pflag.Flag
	Resolve(ctx *Context, flags *pflag.FlagSet) error
}

type FlagResolveFunc func(name string, flags *pflag.FlagSet) error

// IdentityProvider represents the interface used to implement an identity provider
// plugin. It provides authentication and authorization functionality.
type IdentityProvider interface {
	Plugin

	// Authenticate will authenticate a user and return details of
	// their identity.
	Authenticate(ctx *Context, clusterProvider string) (Identity, error)
}

// IdentityStore represents an way to store and retrieve credentials
type IdentityStore interface {
	CredsExists() (bool, error)
	Save(identity Identity) error
	Load() (Identity, error)
	Expired() bool
}

// DiscoverOutput holds details of the output of the discovery
type DiscoverOutput struct {
	ClusterProviderName  string `yaml:"clusterprovider"`
	IdentityProviderName string `yaml:"identityprovider"`

	Clusters map[string]*Cluster `yaml:"clusters"`
}

// Cluster represents the information about a discovered k8s cluster
type Cluster struct {
	ID                       string  `yaml:"id"`
	Name                     string  `yaml:"name"`
	ControlPlaneEndpoint     *string `yaml:"endpoint"`
	CertificateAuthorityData *string `yaml:"ca"`
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
