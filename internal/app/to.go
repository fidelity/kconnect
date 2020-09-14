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

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
)

type ConnectToParams struct {
	CommonConfig
	HistoryConfig
	KubernetesConfig

	Alias    string `json:"alias"`
	EntryID  string `json:"id"`
	Password string `json:"password"`

	Context *provider.Context
}

func (a *App) ConnectTo(params *ConnectToParams) error {
	a.logger.Debug("Running ConnectTo")

	entry, err := a.getHistoryEntry(params.EntryID, params.Alias)
	if err != nil {
		return fmt.Errorf("getting history entry: %w", err)
	}

	idProvider, err := provider.GetIdentityProvider(entry.Spec.Identity)
	if err != nil {
		return fmt.Errorf("getting identity provider %s: %w", entry.Spec.Identity, err)
	}
	clusterProvider, err := provider.GetClusterProvider(entry.Spec.Provider)
	if err != nil {
		return fmt.Errorf("getting cluster provider %s: %w", entry.Spec.Provider, err)
	}

	cs, err := a.buildConnectToConfig(idProvider, clusterProvider, entry)
	if err != nil {
		return fmt.Errorf("building connectTo config set: %w", err)
	}

	if params.Password != "" {
		if err := cs.SetValue("password", params.Password); err != nil {
			return fmt.Errorf("setting password config item: %w", err)
		}
	}

	ctx := provider.NewContext(
		provider.WithConfig(cs),
		provider.WithLogger(a.logger),
	)

	useParams := &UseParams{
		Context:          ctx,
		IdentityProvider: idProvider,
		Provider:         clusterProvider,
	}

	if err := config.Unmarshall(cs, useParams); err != nil {
		return fmt.Errorf("unmarshalling config into use params: %w", err)
	}

	useParams.NoHistory = true
	useParams.ClusterID = &entry.Spec.ProviderID

	identity, err := useParams.IdentityProvider.Authenticate(useParams.Context, useParams.Provider.Name())
	if err != nil {
		return fmt.Errorf("authenticating using provider %s: %w", useParams.IdentityProvider.Name(), err)
	}
	useParams.Identity = identity

	return a.Use(useParams)
}

func (a *App) getHistoryEntry(id, alias string) (*historyv1alpha.HistoryEntry, error) {
	if id != "" {
		return a.historyStore.GetByID(id)
	}

	//TODO: should we only allow 1 alias
	entries, err := a.historyStore.GetByAlias(alias)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, nil
	}

	return entries[0], nil
}

func (a *App) buildConnectToConfig(idProvider provider.IdentityProvider, clusterProvider provider.ClusterProvider, historyEntry *historyv1alpha.HistoryEntry) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()
	if err := cs.AddSet(idProvider.ConfigurationItems()); err != nil {
		return nil, fmt.Errorf("adding identity provider config items: %w", err)
	}
	if err := cs.AddSet(clusterProvider.ConfigurationItems()); err != nil {
		return nil, fmt.Errorf("adding cluster provider config items: %w", err)
	}
	if err := provider.AddCommonIdentityConfig(cs); err != nil {
		return nil, fmt.Errorf("adding common identity config items: %w", err)
	}
	if err := provider.AddCommonClusterConfig(cs); err != nil {
		return nil, fmt.Errorf("adding common cluster config items: %w", err)
	}
	if err := AddKubeconfigConfigItems(cs); err != nil {
		return nil, fmt.Errorf("adding kubeconfig config items: %w", err)
	}

	for k, v := range historyEntry.Spec.Flags {
		configItem := cs.Get(k)
		if configItem == nil {
			a.logger.Debugf("no config item called %s found", k)
			continue
		}

		if configItem.Type == config.ItemTypeString {
			configItem.Value = v
		} else if configItem.Type == config.ItemTypeInt {
			intVal, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			configItem.Value = intVal
		} else if configItem.Type == config.ItemTypeBool {
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			configItem.Value = boolVal
		} else {
			return nil, fmt.Errorf("trying to set config item %s of type %s: %w", configItem.Name, configItem.Type, ErrUnknownConfigItemType)
		}
	}

	return cs, nil
}
