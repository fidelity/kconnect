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
	"errors"
	"fmt"
	"strings"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"go.uber.org/zap"
	"gopkg.in/ini.v1"

	"github.com/fidelity/kconnect/pkg/aws/awsconfig"
	"github.com/fidelity/kconnect/pkg/history"
)

const (
	EKSProviderName     = "eks"
	AKSProviderName     = "aks"
	RancherProviderName = "rancher"
	OIDCProviderName    = "oidc"
)

type LogoutInput struct {
	CommonConfig
	HistoryConfig
	KubernetesConfig

	All   bool
	Alias string
	IDs   string
}

func (a *App) Logout(ctx context.Context, params *LogoutInput) error {
	entries, err := a.getClustersToLogout(params)
	if err != nil {
		return err
	}

	if entries == nil || len(entries.Items) == 0 {
		return ErrNoEntriesFound
	}

	for i := range entries.Items {
		err = a.doLogout(params, &entries.Items[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) getClustersToLogout(params *LogoutInput) (*historyv1alpha.HistoryEntryList, error) {
	entries := historyv1alpha.NewHistoryEntryList()

	var err error

	switch {
	case params.All:
		zap.S().Infof("will log out of all clusters")

		entries, err = a.historyStore.GetAllSortedByLastUsed()
		if err != nil {
			return nil, err
		}
	case params.Alias != "" || params.IDs != "":
		// Aliases and ids can both be supplied, in which case all of the
		// entries they refer to are logged out of.
		for _, alias := range splitCommaList(params.Alias) {
			entry, err := a.historyStore.GetByAlias(alias)
			if err != nil {
				return nil, err
			}

			if entry == nil {
				return nil, fmt.Errorf("%w: %s", ErrAliasNotFound, alias)
			}

			entries.Items = append(entries.Items, *entry)
		}

		for _, id := range splitCommaList(params.IDs) {
			entry, err := a.historyStore.GetByID(id)
			if err != nil {
				return nil, err
			}

			if entry == nil {
				return nil, fmt.Errorf("%w: %s", history.ErrEntryNotFound, id)
			}

			entries.Items = append(entries.Items, *entry)
		}
	default:
		zap.S().Infof("Logging out of current cluster")

		entry, err := a.historyStore.GetLastModified(0)
		if err != nil {
			return nil, err
		}

		entries.Items = append(entries.Items, *entry)
	}

	return entries, nil
}

// splitCommaList splits a comma separated list, ignoring surrounding whitespace
// and any empty items.
func splitCommaList(list string) []string {
	items := []string{}

	for item := range strings.SplitSeq(list, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}

	return items
}

// entryAlias returns the alias of a history entry, which is optional.
func entryAlias(entry *historyv1alpha.HistoryEntry) string {
	if entry.Spec.Alias == nil {
		return ""
	}

	return *entry.Spec.Alias
}

func (a *App) doLogout(params *LogoutInput, entry *historyv1alpha.HistoryEntry) error {
	switch entry.Spec.Provider {
	case EKSProviderName:
		err := a.doLogoutEKS(entry)
		if err != nil {
			return err
		}
	case AKSProviderName, RancherProviderName, OIDCProviderName:
		err := a.doLogoutKubeconfigUser(params, entry)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnknownProvider, entry.Spec.Provider)
	}

	return nil
}

func (a *App) doLogoutEKS(entry *historyv1alpha.HistoryEntry) error {
	zap.S().Infof("logging out of entry (eks): name: %s, alias: %s", entry.Name, entryAlias(entry))

	profileName, ok := entry.Spec.Flags["aws-profile"]
	if !ok {
		zap.S().Infof("no aws profile name found for entry %s", entry.Name)
		return nil
	}

	path, err := awsconfig.LocateConfigFile()
	if err != nil {
		return err
	}

	cfg, err := ini.Load(path)
	if err != nil {
		return err
	}

	cfg.DeleteSection(profileName)

	return cfg.SaveTo(path)
}

// doLogoutKubeconfigUser logs out of the providers whose credentials are held
// by the user in the kubeconfig, which is removed.
func (a *App) doLogoutKubeconfigUser(params *LogoutInput, entry *historyv1alpha.HistoryEntry) error {
	zap.S().Infof("logging out of entry (%s): name: %s, alias: %s", entry.Spec.Provider, entry.Name, entryAlias(entry))
	return a.deleteUserFromKubeconfigByEntryID(params.Kubeconfig, entry.Name)
}

func (a *App) deleteUserFromKubeconfigByEntryID(kubeconfigPath, entryID string) error {
	config, err := kubeconfig.Read(kubeconfigPath)
	if err != nil {
		return err
	}

	kubeconfigUser := ""

	for context := range config.Contexts {
		historyRef, err := historyv1alpha.GetHistoryReferenceFromContext(config.Contexts[context])
		if err != nil {
			// A kubeconfig will usually contain contexts that were not
			// created by kconnect, which are of no interest here.
			if errors.Is(err, historyv1alpha.ErrNoHistoryExtension) {
				continue
			}

			return err
		}

		if historyRef.EntryID == entryID {
			kubeconfigUser = config.Contexts[context].AuthInfo
			break
		}
	}

	if kubeconfigUser == "" {
		zap.S().Infof("no user found in kubeconfig for entry: %s", entryID)
		return nil
	}

	delete(config.AuthInfos, kubeconfigUser)

	return kubeconfig.Write(kubeconfigPath, config, false, false)
}
