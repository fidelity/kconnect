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
	Context *provider.Context
}

func (a *App) Logout(params *LogoutParams) error {

	fmt.Println("Logging out!!")
	// if no --alias, or --all .. log out of current
	entries := historyv1alpha.NewHistoryEntryList()
	//entries.Items = []historyv1alpha.HistoryEntry{}
	var  err error
	if params.All {
		fmt.Println("Getting all")
		entries, err = a.historyStore.GetAllSortedByLastUsed()
		if err != nil {
			return err
		}
	} else if params.Alias != "" {
		fmt.Printf("Alias! %s\n", params.Alias)
		aliasList := strings.Split(params.Alias, ",")
		for _, alias := range(aliasList) {
			entry, err := a.historyStore.GetByAlias(alias)
			if err != nil {
				return err
			}
			if entry == nil {
				return ErrAliasNotFound
			}
			fmt.Printf("entry: %+v\n", entry)
			entries.Items = append(entries.Items, *entry)
		}
	} else {
		// Log out of current cluster
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

	fmt.Printf("logging out of entry: %s\n", entry.Name)
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
	config, err := kubeconfig.Read(params.Kubeconfig)
	if err != nil {
		return err
	}

	config.AuthInfos[]
	
	fmt.Printf("logging out of entry: %s\n", entry.Name)
	return nil
}