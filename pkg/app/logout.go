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
	"strings"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"go.uber.org/zap"
	"gopkg.in/ini.v1"

	"github.com/fidelity/kconnect/pkg/aws/awsconfig"
)

const (
	EKSProviderName     = "eks"
	AKSProviderName     = "aks"
	RancherProviderName = "rancher"
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
	case params.Alias != "":
		aliasList := strings.Split(params.Alias, ",")
		for _, alias := range aliasList {
			entry, err := a.historyStore.GetByAlias(alias)
			if err != nil {
				return nil, err
			}

			if entry == nil {
				return nil, ErrAliasNotFound
			}

			entries.Items = append(entries.Items, *entry)
		}

		fallthrough
	case params.IDs != "":
		idsList := strings.Split(params.IDs, ",")
		for _, id := range idsList {
			entry, err := a.historyStore.GetByID(id)
			if err != nil {
				return nil, err
			}

			if entry == nil {
				return nil, ErrAliasNotFound
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

func (a *App) doLogout(params *LogoutInput, entry *historyv1alpha.HistoryEntry) error {
	// TODO use const for provider
	switch entry.Spec.Provider {
	case EKSProviderName:
		err := a.doLogoutEKS(entry)
		if err != nil {
			return err
		}
	case AKSProviderName:
		err := a.doLogoutAKS(params, entry)
		if err != nil {
			return err
		}
	case RancherProviderName:
		err := a.doLogoutRancher(params, entry)
		if err != nil {
			return err
		}
	default:
		return ErrUnknownProvider
	}

	return nil
}

func (a *App) doLogoutEKS(entry *historyv1alpha.HistoryEntry) error {
	zap.S().Infof("logging out of entry (eks): name: %s, alias: %s", entry.Name, *entry.Spec.Alias)

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

func (a *App) doLogoutAKS(params *LogoutInput, entry *historyv1alpha.HistoryEntry) error {
	zap.S().Infof("logging out of entry (aks): name: %s, alias: %s", entry.Name, *entry.Spec.Alias)
	return a.deleteUserFromKubeconfigByEntryID(params.Kubeconfig, entry.Name)
}

func (a *App) doLogoutRancher(params *LogoutInput, entry *historyv1alpha.HistoryEntry) error {
	zap.S().Infof("logging out of entry (rancher): name: %s, alias: %s", entry.Name, *entry.Spec.Alias)
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
		if err != nil && errors.Is(err, historyv1alpha.ErrNoHistoryExtension) {
			return err
		}

		if historyRef.EntryID == entryID {
			kubeconfigUser = config.Contexts[context].AuthInfo
			break
		}
	}

	if kubeconfigUser == "" {
		zap.S().Infof("no user found in kubeconfig for entry: %S", entryID)
		return nil
	}

	delete(config.AuthInfos, kubeconfigUser)

	return kubeconfig.Write(kubeconfigPath, config, false, false)
}
