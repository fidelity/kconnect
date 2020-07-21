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
	"fmt"
	"sync"
)

var (
	pluginsLock     sync.Mutex
	identityPlugins = make(map[string]IdentityProvider)
	clusterPlugins  = make(map[string]ClusterProvider)
)

func RegisterIdentityProviderPlugin(name string, plugin IdentityProvider) error {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if _, found := identityPlugins[name]; found {
		return fmt.Errorf("Identity plugin with name %q already registsred", name)
	}
	identityPlugins[name] = plugin

	return nil
}

func RegisterClusterProviderPlugin(name string, plugin ClusterProvider) error {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if _, found := clusterPlugins[name]; found {
		return fmt.Errorf("Cluster plugin with name %q already registsred", name)
	}
	clusterPlugins[name] = plugin

	return nil
}

func GetClusterProvider(name string) (ClusterProvider, error) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if plugin, found := clusterPlugins[name]; found {
		return plugin, nil
	}

	return nil, fmt.Errorf("no cluster plugin with name %q", name)
}

func GetIdentityProvider(name string) (IdentityProvider, error) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	if plugin, found := identityPlugins[name]; found {
		return plugin, nil
	}

	return nil, fmt.Errorf("no identity plugin with name %q", name)
}

func ListClusterProviders() []ClusterProvider {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	plugins := make([]ClusterProvider, 0)
	for _, value := range clusterPlugins {
		plugins = append(plugins, value)
	}

	return plugins
}

func ListIdentityProviders() []IdentityProvider {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	plugins := make([]IdentityProvider, 0)
	for _, value := range identityPlugins {
		plugins = append(plugins, value)
	}

	return plugins
}
