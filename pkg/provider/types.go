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
	"github.com/fidelity/kconnect/pkg/config"
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

	// Get will get the details of a cluster with the provided provider
	// specific cluster id
	Get(ctx *Context, clusterID string, identity Identity) (*Cluster, error)

	// GetClusterConfig will get the kubeconfig for a cluster
	GetClusterConfig(ctx *Context, cluster *Cluster, namespace string) (*api.Config, string, error)

	// ConfigurationResolver returns the resolver used to interactively resolve configuration
	ConfigurationResolver() ConfigResolver
}

// ConfigResolver is used to resolve the values for config items interactively.
type ConfigResolver interface {

	// Validate is used to validate the config items and return any errors
	Validate(config config.ConfigurationSet) error

	// Resolve will resolve the values for the supplied config items. It will interactively
	// resolve the values by asking the user for selections.
	Resolve(config config.ConfigurationSet) error
}

type ConfigResolveFunc func(name string, config *config.ConfigurationSet) error

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

	// ConfigurationItems will return the configuration items for the plugin
	ConfigurationItems() config.ConfigurationSet

	// Usage returns a string to display for help
	Usage() string
}
