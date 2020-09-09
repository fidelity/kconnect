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

	"k8s.io/apimachinery/pkg/runtime"

	historyv1alpha "github.com/fidelity/kconnect/pkg/history/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"github.com/fidelity/kconnect/pkg/provider"
)

// UseParams are the parameters to the use function
type UseParams struct {
	HistoryConfig
	KubernetesConfig
	provider.IdentityProviderConfig

	SetCurrent bool `json:"set-current,omitempty"`

	Provider         provider.ClusterProvider
	IdentityProvider provider.IdentityProvider
	Identity         provider.Identity
	Context          *provider.Context

	Args map[string]string
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

	entry := historyv1alpha.NewHistoryEntry()
	//TODO - move this to a function
	//entry.Spec.Alias = ""
	entry.Spec.ConfigFile = params.Kubeconfig
	entry.Spec.Flags = params.Args //TODO: filter sensitive flags
	entry.Spec.Identity = params.IdentityProvider.Name()
	entry.Spec.Provider = params.Provider.Name()
	entry.Spec.ProviderID = cluster.ID

	if err := a.historyStore.Add(entry); err != nil {
		return fmt.Errorf("adding connection to history: %w", err)
	}

	historyRef := historyv1alpha.NewHistoryReference(entry.Spec.ID)
	kubeConfig.Contexts[contextName].Extensions = make(map[string]runtime.Object)
	kubeConfig.Contexts[contextName].Extensions["kconnect"] = historyRef

	if err := kubeconfig.Write(params.Kubeconfig, kubeConfig); err != nil {
		return fmt.Errorf("writing cluster kubeconfig: %w", err)
	}

	return nil
}
