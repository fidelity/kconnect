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

// IdentityProvider represents the interface used to implement an identity provider
// plugin. It provides authentication and authorization functionality.
// Initial plugins will be SAML, OAuth/OIDC, AzureAD
type IdentityProvider interface {
	// Name returns the name of the plugin
	Name() string

	// Flags will return the flags as part of a flagset for the plugin
	Flags() *pflag.FlagSet

	// Authenticate will authenticate a user and return details of
	// their identity.
	Authenticate() (*Identity, error)
}

// NewIdentityProvider is a factory method for creating an identity provider
// based on name
func NewIdentityProvider(name string) *IdentityProvider {
	return nil
}

type SAMLIdentityProvider struct{}
type OIDCIdentityprovider struct{}
type AzureADIdentityProvider struct{}

// Provider is the interface that is used to implement providers
// of Kubernetes clusters. There will be implementations for AWS
// Azure and Rancher initially. Its the job of the provider to
// discover clusters based on a users identity and to generate
// a kubeconfig (and any other files) that are required to access
// a selected cluster.
type ClusterProvider interface {
	// Flags returns the list of CLI flags that are specific to the
	// the provider
	Flags() *pflag.FlagSet

	// SupportedAuthorizers returns a list of authorizers thats this
	// provider supports.
	// NOTE: could be separate
	SupportedIdentityProviders() []string

	// Discover will return a list of clusters for the user using the supplied identuty
	Discover(identity interface{}, flags map[string]string) ([]*Cluster, error)

	// Generate will generate the kubeconfig and any other associated files for
	// connection to a cluster with a specific identifier
	Generate(clusterId string) error
}

// Sample cluster providers
type AKSClusterProvider struct{}
type EKSClusterProvider struct{}
type RancherClusterprovider struct{}

// FlagsResolver is used to resolve the values for flags interactively.
// There will be a flags resolver for Azure, AWS and Rancher initially.
type FlagsResolver interface {

	// Resolve will resolve the values for the supplied flags. It would interactively
	// resolve the values by asking the user for selections. It will basically set the
	// Value on each pflag.Flag
	Resolve(identity interface{}, flags []*pflag.Flag) error
}

type AzureFlagsResolver struct{}
type AWSFlagsResolver struct{}
type RancherFlagsResolver struct{}

// Cluster represents the information about a discovered k8s cluster
// NOTE: fields in this struct are only for illustration and more though needs to
// go into it
type Cluster interface {
	Name() string

	ID() string

	Region() string
}
