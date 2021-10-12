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
	"context"
	"fmt"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
)

// UseInput are the parameters to the use function
type UseInput struct {
	CommonConfig
	CommonUseConfig
	HistoryConfig
	KubernetesConfig
	common.IdentityProviderConfig
	common.ClusterProviderConfig

	SetCurrent bool `json:"set-current,omitempty"`

	DiscoveryProvider string
	IdentityProvider  string

	IgnoreAlias bool

	ConfigSet config.ConfigurationSet
}

func (a *App) Use(ctx context.Context, input *UseInput) error {
	a.logger.Debug("use command")
	identityProvider, err := a.getIdentityProvider(&input.IdentityProvider, &input.DiscoveryProvider)
	if err != nil {
		return fmt.Errorf("getting identity provider: %w", err)
	}
	clusterProvider, err := a.getDiscoveryProvider(&input.DiscoveryProvider, &input.IdentityProvider)
	if err != nil {
		return fmt.Errorf("getting discovery provider: %w", err)
	}

	if !isIdpSupported(identityProvider.Name(), clusterProvider) {
		return fmt.Errorf("using identity provider %s: %w", input.IdentityProvider, ErrUnsuportedIdpProtocol)
	}

	err = clusterProvider.CheckPreReqs()
	if err != nil {
		//TODO: how to report this???
		fmt.Fprintf(os.Stderr, "\033[33m%s\033[0m\n", err.Error())
	}

	authOutput, err := identityProvider.Authenticate(ctx, &identity.AuthenticateInput{
		ConfigSet: input.ConfigSet,
	})
	if err != nil {
		return fmt.Errorf("authenticating using provider %s: %w", identityProvider.Name(), err)
	}

	if err := clusterProvider.Resolve(input.ConfigSet, authOutput.Identity); err != nil {
		return fmt.Errorf("resolving config items: %w", err)
	}

	var cluster *discovery.Cluster
	if input.ClusterID == nil || *input.ClusterID == "" {
		cluster, err = a.discoverCluster(ctx, clusterProvider, authOutput.Identity, input)
	} else {
		cluster, err = a.getCluster(ctx, clusterProvider, authOutput.Identity, input)
	}
	if err != nil {
		return err
	}
	if cluster == nil {
		return nil
	}

	output, err := clusterProvider.GetConfig(ctx, &discovery.GetConfigInput{
		Cluster:   cluster,
		Namespace: &input.Namespace,
		Identity:  authOutput.Identity,
	})
	if err != nil {
		return fmt.Errorf("creating kubeconfig for %s: %w", cluster.Name, err)
	}

	if !input.IgnoreAlias {
		if err := a.resolveAndCheckAlias(input); err != nil {
			return fmt.Errorf("resolving and checking alias: %w", err)
		}
	}

	historyID := input.EntryID
	if !input.NoHistory {
		entry := historyv1alpha.NewHistoryEntry()
		entry.Spec.Alias = input.Alias
		entry.Spec.ConfigFile = input.Kubeconfig
		entry.Spec.Flags = a.filterConfig(input)
		entry.Spec.Identity = input.IdentityProvider
		entry.Spec.Provider = input.DiscoveryProvider
		entry.Spec.ProviderID = cluster.ID

		if err := a.historyStore.Add(entry); err != nil {
			return fmt.Errorf("adding connection to history: %w", err)
		}

		historyID = entry.ObjectMeta.Name
	}

	kubeConfig := output.KubeConfig
	contextName := *output.ContextName
	if historyID != "" {
		historyRef := historyv1alpha.NewHistoryReference(historyID)
		kubeConfig.Contexts[contextName].Extensions = make(map[string]runtime.Object)
		kubeConfig.Contexts[contextName].Extensions["kconnect"] = historyRef
	}

	if input.Kubeconfig == "" {
		a.logger.Debug("no kubeconfig supplied, setting default")
		pathOptions := clientcmd.NewDefaultPathOptions()
		input.Kubeconfig = pathOptions.GetDefaultFilename()
	}

	if err := kubeconfig.Write(input.Kubeconfig, kubeConfig, true, input.SetCurrent); err != nil {
		return fmt.Errorf("writing cluster kubeconfig: %w", err)
	}

	return nil
}

func (a *App) discoverCluster(ctx context.Context, clusterProvider discovery.Provider, identity identity.Identity, params *UseInput) (*discovery.Cluster, error) {
	a.logger.Infow("discovering clusters", "provider", params.DiscoveryProvider)

	discoverOutput, err := clusterProvider.Discover(ctx, &discovery.DiscoverInput{
		ConfigSet: params.ConfigSet,
		Identity:  identity,
	})
	if err != nil {
		return nil, fmt.Errorf("discovering clusters using %s: %w", clusterProvider.Name(), err)
	}

	if discoverOutput.Clusters == nil || len(discoverOutput.Clusters) == 0 {
		a.logger.Warn("no clusters discovered")
		return nil, nil
	}

	cluster, err := a.selectCluster(discoverOutput)
	if err != nil {
		return nil, fmt.Errorf("selecting cluster: %w", err)
	}

	return cluster, nil
}

func (a *App) getCluster(ctx context.Context, clusterProvider discovery.Provider, identity identity.Identity, params *UseInput) (*discovery.Cluster, error) {
	a.logger.Infow("getting cluster details", "id", *params.ClusterID, "provider", params.DiscoveryProvider)

	output, err := clusterProvider.GetCluster(ctx, &discovery.GetClusterInput{
		ClusterID: *params.ClusterID,
		Identity:  identity,
		ConfigSet: params.ConfigSet,
	})
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}
	if output.Cluster == nil {
		return nil, fmt.Errorf("getting cluster with id %s: %w", *params.ClusterID, ErrClusterNotFound)
	}

	return output.Cluster, nil
}

func (a *App) filterConfig(params *UseInput) map[string]string {
	filteredConfig := make(map[string]string)

	for _, configItem := range params.ConfigSet.GetAll() {

		if configItem.Sensitive || configItem.HistoryIgnore {
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

func (a *App) resolveAndCheckAlias(params *UseInput) error {
	if params.Alias == nil || *params.Alias == "" {
		if !a.interactive {
			return nil
		}
		a.logger.Debug("no alias set, resolving")
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

	a.logger.Infof("Command to reconnect using this alias: kconnect to %s", *params.Alias)

	return nil
}

func (a *App) resolveAlias() (string, error) {
	useAlias, err := prompt.Confirm("use-alias", "Do you want to set an alias?", false)
	if err != nil {
		return "", err
	}

	if !useAlias {
		return "", nil
	}

	alias, err := prompt.Input("alias", "Enter the alias name", false)
	if err != nil {
		return "", err
	}

	return alias, nil
}

func (a *App) getDiscoveryProvider(name *string, scopedToIdentityProvider *string) (discovery.Provider, error) {
	if name == nil || *name == "" {
		return nil, ErrDiscoveryProviderRequired
	}

	prov, err := registry.GetDiscoveryProvider(*name, &provider.PluginCreationInput{
		Logger:        a.logger.With("provider", name),
		IsInteractice: a.interactive,
		ItemSelector:  a.itemSelector,
		ScopedTo:      scopedToIdentityProvider,
		HTTPClient:    a.httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("getting discovery provider %s: %w", *name, err)
	}

	return prov, nil
}

func (a *App) getIdentityProvider(name *string, scopedToDiscoveryProvider *string) (identity.Provider, error) {
	if name == nil || *name == "" {
		return nil, ErrIdentityProviderRequired
	}

	prov, err := registry.GetIdentityProvider(*name, &provider.PluginCreationInput{
		Logger:        a.logger.With("provider", name),
		IsInteractice: a.interactive,
		ItemSelector:  a.itemSelector,
		ScopedTo:      scopedToDiscoveryProvider,
		HTTPClient:    a.httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("getting identity provider %s: %w", *name, err)
	}

	return prov, nil
}

func isIdpSupported(idProviderName string, clusterProvider discovery.Provider) bool { //nolint: interfacer
	registration, _ := registry.GetDiscoveryProviderRegistration(clusterProvider.Name())
	supportedIDProviders := registration.SupportedIdentityProviders

	for _, supportedIDProvider := range supportedIDProviders {
		if supportedIDProvider == idProviderName {
			return true
		}
	}

	return false
}
