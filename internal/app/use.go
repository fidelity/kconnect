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
	provider.ClusterProviderConfig

	SetCurrent bool `json:"set-current,omitempty"`

	Provider         provider.ClusterProvider
	IdentityProvider provider.IdentityProvider
	Identity         provider.Identity

	Context *provider.Context
}

func (a *App) Use(params *UseParams) error {
	a.logger.Debug("use command")
	clusterProvider := params.Provider

	var cluster *provider.Cluster
	var err error

	aliasInUse, err := a.aliasInUse(params.Alias)
	if err != nil {
		return err
	}
	if aliasInUse {
		return ErrAliasAlreadyUsed
	}

	if params.ClusterID == nil || *params.ClusterID == "" {
		cluster, err = a.discoverCluster(params)
	} else {
		cluster, err = a.getCluster(params)
	}
	if err != nil {
		return err
	}
	if cluster == nil {
		return nil
	}

	kubeConfig, contextName, err := clusterProvider.GetClusterConfig(params.Context, cluster)
	if err != nil {
		return fmt.Errorf("creating kubeconfig for %s: %w", cluster.Name, err)
	}

	historyID := params.EntryID
	if !params.NoHistory {
		entry := historyv1alpha.NewHistoryEntry()
		entry.Spec.Alias = params.Alias
		entry.Spec.ConfigFile = params.Kubeconfig
		entry.Spec.Flags = a.filterConfig(params)
		entry.Spec.Identity = params.IdentityProvider.Name()
		entry.Spec.Provider = params.Provider.Name()
		entry.Spec.ProviderID = cluster.ID

		if err := a.historyStore.Add(entry); err != nil {
			return fmt.Errorf("adding connection to history: %w", err)
		}

		historyID = entry.ObjectMeta.Name
	}

	if historyID != "" {
		historyRef := historyv1alpha.NewHistoryReference(historyID)
		kubeConfig.Contexts[contextName].Extensions = make(map[string]runtime.Object)
		kubeConfig.Contexts[contextName].Extensions["kconnect"] = historyRef
	}

	if err := kubeconfig.Write(params.Kubeconfig, kubeConfig, params.SetCurrent); err != nil {
		return fmt.Errorf("writing cluster kubeconfig: %w", err)
	}

	return nil
}

func (a *App) discoverCluster(params *UseParams) (*provider.Cluster, error) {
	a.logger.Infof("discovering clusters using %s provider", params.Provider.Name())

	clusterProvider := params.Provider
	discoverOutput, err := clusterProvider.Discover(params.Context, params.Identity)
	if err != nil {
		return nil, fmt.Errorf("discovering clusters using %s: %w", clusterProvider.Name(), err)
	}

	if discoverOutput.Clusters == nil || len(discoverOutput.Clusters) == 0 {
		a.logger.Info("no clusters discovered")
		return nil, nil
	}

	cluster, err := a.selectCluster(discoverOutput)
	if err != nil {
		return nil, fmt.Errorf("selecting cluster: %w", err)
	}

	return cluster, nil
}

func (a *App) getCluster(params *UseParams) (*provider.Cluster, error) {
	a.logger.Infof("getting cluster %s using %s provider", *params.ClusterID, params.Provider.Name())

	clusterProvider := params.Provider
	cluster, err := clusterProvider.Get(params.Context, *params.ClusterID, params.Identity)
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}
	if cluster == nil {
		return nil, fmt.Errorf("getting cluster with id %s: %w", *params.ClusterID, ErrClusterNotFound)
	}

	return cluster, nil
}

func (a *App) filterConfig(params *UseParams) map[string]string {
	filteredConfig := make(map[string]string)

	idConfigSet := params.IdentityProvider.ConfigurationItems()
	discConfigSet := params.Provider.ConfigurationItems()
	commonIDConfigSet := provider.CommonIdentityConfig()

	for _, configItem := range params.Context.ConfigurationItems().GetAll() {
		cmnConfig := commonIDConfigSet.Get(configItem.Name)
		idConfig := idConfigSet.Get(configItem.Name)
		discConfig := discConfigSet.Get(configItem.Name)

		if cmnConfig == nil && idConfig == nil && discConfig == nil {
			continue
		}

		if configItem.Sensitive {
			continue
		}

		val := ""
		switch configItem.Type {
		case config.ItemTypeString:
			val = configItem.Value.(string)
		case config.ItemTypeBool:
			boolVal := configItem.Value.(bool)
			val = strconv.FormatBool(boolVal)
		case config.ItemTypeInt:
			intVal := configItem.Value.(int64)
			val = strconv.FormatInt(intVal, 10)
		}

		filteredConfig[configItem.Name] = val
	}

	return filteredConfig
}

func (a *App) aliasInUse(alias *string) (bool, error) {
	if alias == nil || *alias == "" {
		return false, nil
	}

	entry, err := a.historyStore.GetByAlias(*alias)
	if err != nil {
		return false, fmt.Errorf("getting history by alias: %w", err)
	}

	inUse := entry != nil

	return inUse, nil
}
