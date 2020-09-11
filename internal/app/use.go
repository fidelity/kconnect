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

package app

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"github.com/fidelity/kconnect/pkg/provider"
)

// UseParams are the parameters to the use function
type UseParams struct {
	CommonConfig
	HistoryConfig
	KubernetesConfig
	provider.IdentityProviderConfig

	SetCurrent bool `json:"set-current,omitempty"`

	Provider         provider.ClusterProvider
	IdentityProvider provider.IdentityProvider
	Identity         provider.Identity

	Context *provider.Context
}

func (a *App) Use(params *UseParams) error {
	a.logger.Debug("Use command")

	provider := params.Provider

	a.logger.Infof("Discovering clusters using %s provider", provider.Name())
	discoverOutput, err := provider.Discover(params.Context, params.Identity)
	if err != nil {
		return fmt.Errorf("discovering clusters using %s: %w", provider.Name(), err)
	}

	if discoverOutput.Clusters == nil || len(discoverOutput.Clusters) == 0 {
		a.logger.Info("no clusters discovered")
		return nil
	}

	cluster, err := a.selectCluster(discoverOutput)
	if err != nil {
		return fmt.Errorf("selecting cluster: %w", err)
	}

	kubeConfig, contextName, err := provider.GetClusterConfig(params.Context, cluster, params.SetCurrent)
	if err != nil {
		return fmt.Errorf("creating kubeconfig for %s: %w", cluster.Name, err)
	}

	if !params.NoHistory {
		entry := historyv1alpha.NewHistoryEntry()
		//TODO - move this to a function
		//entry.Spec.Alias = ""
		entry.Spec.ConfigFile = params.Kubeconfig
		entry.Spec.Flags = a.filterConfig(params)
		entry.Spec.Identity = params.IdentityProvider.Name()
		entry.Spec.Provider = params.Provider.Name()
		entry.Spec.ProviderID = cluster.ID

		if err := a.historyStore.Add(entry); err != nil {
			return fmt.Errorf("adding connection to history: %w", err)
		}

		historyRef := historyv1alpha.NewHistoryReference(entry.Spec.ID)
		kubeConfig.Contexts[contextName].Extensions = make(map[string]runtime.Object)
		kubeConfig.Contexts[contextName].Extensions["kconnect"] = historyRef
	}

	if err := kubeconfig.Write(params.Kubeconfig, kubeConfig); err != nil {
		return fmt.Errorf("writing cluster kubeconfig: %w", err)
	}

	return nil
}

func (a *App) filterConfig(params *UseParams) map[string]string {
	filteredConfig := make(map[string]string)

	idConfigSet := params.IdentityProvider.ConfigurationItems()
	discConfigSet := params.Provider.ConfigurationItems()
	commonIdConfigSet := provider.CommonIdentityConfig()

	for _, configItem := range params.Context.ConfigurationItems().GetAll() {
		cmnConfig := commonIdConfigSet.Get(configItem.Name)
		idConfig := idConfigSet.Get(configItem.Name)
		discConfig := discConfigSet.Get(configItem.Name)

		if cmnConfig == nil && idConfig == nil && discConfig == nil {
			continue
		}

		if configItem.Sensitive {
			continue
		}

		val := ""
		if configItem.Type == config.ItemTypeString {
			val = configItem.Value.(string)
		} else if configItem.Type == config.ItemTypeBool {
			boolVal := configItem.Value.(bool)
			val = strconv.FormatBool(boolVal)
		} else if configItem.Type == config.ItemTypeInt {
			intVal := configItem.Value.(int64)
			val = strconv.FormatInt(intVal, 10)
		}

		filteredConfig[configItem.Name] = val
	}

	return filteredConfig
}
