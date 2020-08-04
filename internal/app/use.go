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

	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"github.com/fidelity/kconnect/pkg/provider"
)

// UseParams are the parameters to the use function
type UseParams struct {
	Kubeconfig       string
	IdpProtocol      string
	Provider         provider.ClusterProvider
	IdentityProvider provider.IdentityProvider
	Identity         provider.Identity
	Context          *provider.Context
}

func (a *App) Use(params *UseParams) error {
	a.logger.Debug("Use command")

	provider := params.Provider

	a.logger.Infof("Disovering clusters using %s provider", provider.Name())
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

	kubeConfig, err := provider.GetClusterConfig(params.Context, cluster)
	if err != nil {
		return fmt.Errorf("creating kubeconfig for %s: %w", cluster.Name, err)
	}

	if err := kubeconfig.Write(params.Kubeconfig, kubeConfig); err != nil {
		return fmt.Errorf("writing cluster kubeconfig: %w", err)
	}

	return nil
}
