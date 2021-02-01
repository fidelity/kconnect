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
	"sync"

	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
)

var (
	ErrDuplicatePlugin = errors.New("plugin already registered with same name")
	ErrPluginNotFound  = errors.New("plugin not found")
)

var (
	pluginsLock      sync.Mutex
	identityPlugins  = make(map[string]*IdentityPluginRegistration)
	discoveryPlugins = make(map[string]*DiscoveryPluginRegistration)
)

type PluginRegistration struct {
	Name                   string
	UsageExample           string
	ConfigurationItemsFunc provider.ConfigurationItemsFunc
}
type DiscoveryPluginRegistration struct {
	PluginRegistration
	SupportedIdentityProviders []string
	CreateFunc                 discovery.ProviderCreatorFun
}

type IdentityPluginRegistration struct {
	PluginRegistration
	CreateFunc identity.ProviderCreatorFun
}

func RegisterIdentityPlugin(registration *IdentityPluginRegistration) error {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if _, found := identityPlugins[registration.Name]; found {
		return ErrDuplicatePlugin
	}
	identityPlugins[registration.Name] = registration

	return nil
}

func RegisterDiscoveryPlugin(registration *DiscoveryPluginRegistration) error {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if _, found := discoveryPlugins[registration.Name]; found {
		return ErrDuplicatePlugin
	}
	discoveryPlugins[registration.Name] = registration

	return nil
}

func GetDiscoveryProvider(name string, input *provider.PluginCreationInput) (discovery.Provider, error) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if registration, found := discoveryPlugins[name]; found {
		return registration.CreateFunc(input)
	}

	return nil, fmt.Errorf("getting cluster plugin %s: %w", name, ErrPluginNotFound)
}

func GetDiscoveryProviderRegistration(name string) (*DiscoveryPluginRegistration, error) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	registration, found := discoveryPlugins[name]
	if !found {
		return nil, ErrPluginNotFound
	}

	return registration, nil
}

func GetIdentityProvider(name string, input *provider.PluginCreationInput) (identity.Provider, error) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if registration, found := identityPlugins[name]; found {
		return registration.CreateFunc(input)
	}

	return nil, fmt.Errorf("getting identity plugin %s: %w", name, ErrPluginNotFound)
}

func GetIdentityProviderRegistration(name string) (*IdentityPluginRegistration, error) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	registration, found := identityPlugins[name]
	if !found {
		return nil, ErrPluginNotFound
	}

	return registration, nil
}

func ListDiscoveryPluginRegistrations() []*DiscoveryPluginRegistration {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	plugins := []*DiscoveryPluginRegistration{}
	for _, registration := range discoveryPlugins {
		pluginRegistration := registration
		plugins = append(plugins, pluginRegistration)
	}

	return plugins
}

func ListIdentityPluginRegistrations() []*IdentityPluginRegistration {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	plugins := []*IdentityPluginRegistration{}
	for _, registration := range identityPlugins {
		pluginRegistration := registration
		plugins = append(plugins, pluginRegistration)
	}

	return plugins
}
