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

	survey "github.com/AlecAivazis/survey/v2"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"github.com/fidelity/kconnect/pkg/provider"
)

// UseParams are the parameters to the use function
type UseParams struct {
	CommonConfig
	CommonUseConfig
	HistoryConfig
	KubernetesConfig
	provider.IdentityProviderConfig
	provider.ClusterProviderConfig

	SetCurrent bool `json:"set-current,omitempty"`

	Provider         provider.ClusterProvider
	IdentityProvider provider.IdentityProvider
	Identity         provider.Identity

	Context *provider.Context

	IgnoreAlias bool
}

func (a *App) Use(params *UseParams) error {
	zap.S().Debug("use command")
	clusterProvider := params.Provider

	var cluster *provider.Cluster
	var err error

	if !params.IgnoreAlias {
		if err := a.resolveAndCheckAlias(params); err != nil {
			return fmt.Errorf("resolving and checking alias: %w", err)
		}
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

	kubeConfig, contextName, err := clusterProvider.GetClusterConfig(params.Context, cluster, params.Namespace)
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

	if params.Kubeconfig == "" {
		zap.S().Debug("no kubeconfig supplied, setting default")
		pathOptions := clientcmd.NewDefaultPathOptions()
		params.Kubeconfig = pathOptions.GetDefaultFilename()
	}

	if err := kubeconfig.Write(params.Kubeconfig, kubeConfig, true, params.SetCurrent); err != nil {
		return fmt.Errorf("writing cluster kubeconfig: %w", err)
	}

	return nil
}

func (a *App) discoverCluster(params *UseParams) (*provider.Cluster, error) {
	zap.S().Infow("discovering clusters", "provider", params.Provider.Name())

	clusterProvider := params.Provider
	discoverOutput, err := clusterProvider.Discover(params.Context, params.Identity)
	if err != nil {
		return nil, fmt.Errorf("discovering clusters using %s: %w", clusterProvider.Name(), err)
	}

	if discoverOutput.Clusters == nil || len(discoverOutput.Clusters) == 0 {
		zap.S().Warn("no clusters discovered")
		return nil, nil
	}

	cluster, err := a.selectCluster(discoverOutput)
	if err != nil {
		return nil, fmt.Errorf("selecting cluster: %w", err)
	}

	return cluster, nil
}

func (a *App) getCluster(params *UseParams) (*provider.Cluster, error) {
	zap.S().Infow("getting cluster details", "id", *params.ClusterID, "provider", params.Provider.Name())

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

	for _, configItem := range params.Context.ConfigurationItems().GetAll() {

		if configItem.Sensitive {
			continue
		}

		if configItem.HistoryIgnore {
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
			intVal := configItem.Value.(int)
			val = strconv.Itoa(intVal)
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

func (a *App) resolveAndCheckAlias(params *UseParams) error {
	if params.Alias == nil || *params.Alias == "" {
		if !params.Context.IsInteractive() {
			return nil
		}
		zap.S().Debug("no alias set, resolving")
		alias, err := a.resolveAlias()
		if err != nil {
			return fmt.Errorf("resolving alias: %w", err)
		}
		params.Alias = &alias
	}

	aliasInUse, err := a.aliasInUse(params.Alias)
	if err != nil {
		return err
	}
	if aliasInUse {
		return ErrAliasAlreadyUsed
	}

	zap.S().Infof("Command to reconnect using this alias: kconnect to %s", *params.Alias)

	return nil
}

func (a *App) resolveAlias() (string, error) {
	useAlias := false
	promptUse := &survey.Confirm{
		Message: "Do you want to set an alias?",
	}
	if err := survey.AskOne(promptUse, &useAlias); err != nil {
		return "", err
	}

	if !useAlias {
		return "", nil
	}

	alias := ""
	promptAliasName := &survey.Input{
		Message: "Enter the alias name",
	}
	if err := survey.AskOne(promptAliasName, &alias); err != nil {
		return "", err
	}

	return alias, nil
}
