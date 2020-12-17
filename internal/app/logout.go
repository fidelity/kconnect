package app

import (
	"fmt"
	"strings"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/k8s/kubeconfig"
	"github.com/fidelity/kconnect/pkg/provider"
	"go.uber.org/zap"
	"gopkg.in/ini.v1"

	"github.com/fidelity/kconnect/pkg/aws/awsconfig"
)

type LogoutParams struct {
	CommonConfig
	HistoryConfig
	KubernetesConfig

	All     bool
	Alias   string
	IDs		string
	Context *provider.Context
}

func (a *App) Logout(params *LogoutParams) error {

	entries := historyv1alpha.NewHistoryEntryList()
	var  err error
	if params.All {
		zap.S().Infof("will log out of all clusters")
		entries, err = a.historyStore.GetAllSortedByLastUsed()
		if err != nil {
			return err
		}
	} else if params.Alias != "" || params.IDs != "" {
		// Log out of Aliass
		if params.Alias != "" {
			aliasList := strings.Split(params.Alias, ",")
			fmt.Printf("list: %+v", aliasList)
			for _, alias := range(aliasList) {
				zap.S().Infof("will log out of cluster 1: %s", alias)
				entry, err := a.historyStore.GetByAlias(alias)
				if err != nil {
					return err
				}
				if entry == nil {
					return ErrAliasNotFound
				}
				entries.Items = append(entries.Items, *entry)
			}
		}
		if params.IDs != "" {
			idsList := strings.Split(params.IDs, ",")
			for _, id := range(idsList) {
				zap.S().Infof("will log out of cluster2 : %s", id)
				entry, err := a.historyStore.GetByID(id)
				if err != nil {
					return err
				}
				if entry == nil {
					return ErrAliasNotFound
				}
				entries.Items = append(entries.Items, *entry)
			}
		}
	} else {
		// Log out of current cluster
		zap.S().Infof("Logging out of current cluster")
		entry, err := a.historyStore.GetLastModified(0)
		if err != nil {
			return err
		}
		entries.Items = append(entries.Items, *entry)
	}
	if entries == nil || len(entries.Items) == 0 {
		return ErrNoEntriesFound
	}
	for _, entry := range(entries.Items) {
		err = a.doLogout(params, &entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func(a *App) doLogout(params *LogoutParams, entry *historyv1alpha.HistoryEntry) error {

	//TODO use const for provider
	if entry.Spec.Provider == "eks" {
		err := a.doLogoutEKS(entry)
		if err != nil {
			return err
		}
	} else if entry.Spec.Provider == "aks" {
		err := a.doLogoutAKS(params, entry)
		if err != nil {
			return err
		}
	} else if entry.Spec.Provider == "rancher" {
	}
	return nil
}

func(a *App) doLogoutEKS(entry *historyv1alpha.HistoryEntry) error {

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

func(a *App) doLogoutAKS(params *LogoutParams, entry *historyv1alpha.HistoryEntry) error {

	zap.S().Infof("logging out of entry (aks): name: %s, alias: %s", entry.Name, *entry.Spec.Alias)
	config, err := kubeconfig.Read(params.Kubeconfig)
	if err != nil {
		return err
	}
	kubeconfigUser := ""
	for context := range(config.Contexts) {
		historyRef, err := historyv1alpha.GetHistoryReferenceFromContext(config.Contexts[context])
		if err != nil {
			return err
		}
		if historyRef.EntryID == entry.Name {
			kubeconfigUser = config.Contexts[context].AuthInfo
			break
		}
	}
	if kubeconfigUser == "" {
		zap.S().Infof("no user found for entry %s", entry.Name)
		return nil
	}
	delete(config.AuthInfos, kubeconfigUser)
	return kubeconfig.Write(params.Kubeconfig, config, false, false)
}