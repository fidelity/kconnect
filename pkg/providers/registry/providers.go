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
package registry

import (
	"errors"
	"fmt"

	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/fidelity/kconnect/pkg/providers/discovery/aws"
	"github.com/fidelity/kconnect/pkg/providers/identity"
)

var (
	errUnknownClusterProvider  = errors.New("unknown cluster provider")
	errUnknownIdentityProvider = errors.New("unknown identity provider")
)

// ProviderFactory is an interface for the factor that creates ClusterProvider
// and IdentityProvider based on name
type ProviderFactory interface {
	CreateClusterProvider(providerName string) (providers.ClusterProvider, error)
	CreateIdentityProvider(providerName string) (providers.IdentityProvider, error)

	ListClusterProviders() []providers.ClusterProvider
	ListIdentityProviders() []providers.IdentityProvider
}

type providerFactory struct {
	clusterProviders  map[string]providers.ClusterProvider
	identityProviders map[string]providers.IdentityProvider
}

// NewProviderFactory creates a new ProviderFactory
func NewProviderFactory() ProviderFactory {
	clusterProviders := make(map[string]providers.ClusterProvider)
	clusterProviders["eks"] = &aws.EKSClusterProvider{}

	identityProviders := make(map[string]providers.IdentityProvider)
	identityProviders["empty"] = &identity.EmptyIdentityProvider{}
	identityProviders["saml"] = &identity.SAMLIdentityProvider{}

	return &providerFactory{
		clusterProviders:  clusterProviders,
		identityProviders: identityProviders,
	}
}

func (f *providerFactory) CreateClusterProvider(providerName string) (providers.ClusterProvider, error) {
	provider := f.clusterProviders[providerName]
	if provider == nil {
		return nil, fmt.Errorf("getting cluster provider %s: %w", providerName, errUnknownClusterProvider)
	}

	return provider, nil
}

func (f *providerFactory) CreateIdentityProvider(providerName string) (providers.IdentityProvider, error) {
	provider := f.identityProviders[providerName]
	if provider == nil {
		return nil, fmt.Errorf("getting identity provider %s: %w", providerName, errUnknownIdentityProvider)
	}

	return provider, nil
}

func (f *providerFactory) ListClusterProviders() []providers.ClusterProvider {
	plugins := make([]providers.ClusterProvider, 0)
	for _, value := range f.clusterProviders {
		plugins = append(plugins, value)
	}

	return plugins
}

func (f *providerFactory) ListIdentityProviders() []providers.IdentityProvider {
	plugins := make([]providers.IdentityProvider, 0)
	for _, value := range f.identityProviders {
		plugins = append(plugins, value)
	}

	return plugins
}
