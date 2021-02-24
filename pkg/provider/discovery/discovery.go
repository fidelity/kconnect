package discovery

import (
	"context"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
	provcfg "github.com/fidelity/kconnect/pkg/provider/config"
	"github.com/fidelity/kconnect/pkg/provider/identity"
)

// Provider is the interface that is used to implement providers
// of Kubernetes clusters. Its the job of the provider to
// discover clusters based on a users identity and to generate
// a kubeconfig (and any other files) that are required to access
// a selected cluster.
type Provider interface {
	provider.Plugin
	provider.PluginPreReqs
	provcfg.Resolver

	// Discover will discover what clusters the supplied identity has access to
	Discover(ctx context.Context, input *DiscoverInput) (*DiscoverOutput, error)

	// Get will get the details of a cluster with the provided provider
	// specific cluster id
	GetCluster(ctx context.Context, input *GetClusterInput) (*GetClusterOutput, error)

	// GetClusterConfig will get the kubeconfig for a cluster
	GetConfig(ctx context.Context, input *GetConfigInput) (*GetConfigOutput, error)
}

type ProviderCreatorFun func(input *provider.PluginCreationInput) (Provider, error)

// DiscoverInput is the input to Discover
type DiscoverInput struct {
	ConfigSet config.ConfigurationSet
	Identity  identity.Identity
}

// DiscoverOutput holds details of the output of the Discover
type DiscoverOutput struct {
	DiscoveryProvider string `yaml:"discoveryProvider"`
	IdentityProvider  string `yaml:"identityProvider"`

	Clusters map[string]*Cluster `yaml:"clusters"`
}

// GetClusterInput is the input to GetCluster
type GetClusterInput struct {
	ClusterID string
	ConfigSet config.ConfigurationSet
	Identity  identity.Identity
}

// GetClusterOutput is the output to GetCluster
type GetClusterOutput struct {
	Cluster *Cluster
}

type GetConfigInput struct {
	Cluster   *Cluster
	Namespace *string
	Identity  identity.Identity
}

type GetConfigOutput struct {
	KubeConfig  *api.Config
	ContextName *string
}

// Cluster represents the information about a discovered k8s cluster
type Cluster struct {
	ID                       string  `yaml:"id"`
	Name                     string  `yaml:"name"`
	ControlPlaneEndpoint     *string `yaml:"endpoint"`
	CertificateAuthorityData *string `yaml:"ca"`
}
