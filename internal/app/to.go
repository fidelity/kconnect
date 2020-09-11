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

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/provider"
)

type ConnectToParams struct {
	CommonConfig
	HistoryConfig
	KubernetesConfig

	Alias   string `flag:"alias"`
	EntryID string `flag:"id"`

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

	idProvider.Authenticate(params.Context, clusterProvider.Name())

	return nil
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
