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
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/printer"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/registry"
)

type ConnectToInput struct {
	CommonConfig
	HistoryConfig
	KubernetesConfig

	AliasOrIDORPosition string
	Password            string `json:"password"`
	SetCurrent          bool   `json:"set-current,omitempty"`
}

func (a *App) ConnectTo(ctx context.Context, params *ConnectToInput) error {
	a.logger.Debug("running connectto")

	entry, err := a.getHistoryEntry(params)
	if err != nil {
		return fmt.Errorf("getting history entry: %w", err)
	}
	if entry == nil {
		return history.ErrEntryNotFound
	}
	historyID := entry.ObjectMeta.Name

	cs, err := a.buildConnectToConfig(params.ConfigFile, entry.Spec.Provider, entry.Spec.Identity, entry)
	if err != nil {
		return fmt.Errorf("building connectTo config set: %w", err)
	}

	if params.Password != "" {
		if err := cs.SetValue("password", params.Password); err != nil {
			return fmt.Errorf("setting password config item: %w", err)
		}
	}

	useParams := &UseInput{
		IdentityProvider:  entry.Spec.Identity,
		DiscoveryProvider: entry.Spec.Provider,
		ConfigSet:         cs,
	}

	if err := config.Unmarshall(cs, useParams); err != nil {
		return fmt.Errorf("unmarshalling config into use params: %w", err)
	}

	useParams.EntryID = historyID
	useParams.ClusterID = &entry.Spec.ProviderID
	useParams.SetCurrent = params.SetCurrent
	useParams.IgnoreAlias = true
	useParams.Alias = entry.Spec.Alias
	useParams.Kubeconfig = params.Kubeconfig

	return a.Use(ctx, useParams)
}

func (a *App) getHistoryEntry(params *ConnectToInput) (*historyv1alpha.HistoryEntry, error) {

	idOrAliasORPosition := params.AliasOrIDORPosition
	if idOrAliasORPosition == "" {
		entry, err := a.getInteractive(params)
		if err != nil {
			return nil, fmt.Errorf("getting history interactivly: %w", err)
		}
		return entry, nil
	}
	if idOrAliasORPosition == "-" || idOrAliasORPosition == "LAST" {
		entry, err := a.historyStore.GetLastModified(0)
		if err != nil {
			return nil, fmt.Errorf("getting history by last modified index: %w", err)
		}
		return entry, nil
	}
	lastPositionRegex := regexp.MustCompile("LAST~[0-9]+")
	getPositionRegex := regexp.MustCompile("[0-9]+")
	if lastPositionRegex.MatchString(idOrAliasORPosition) {
		n, err := strconv.Atoi(getPositionRegex.FindString(idOrAliasORPosition))
		if err != nil {
			return nil, err
		}
		entry, err := a.historyStore.GetLastModified(n)
		if err != nil {
			return nil, fmt.Errorf("getting history by last modified index: %w", err)
		}
		return entry, nil
	}

	entry, err := a.historyStore.GetByID(idOrAliasORPosition)
	if err != nil {
		return nil, fmt.Errorf("getting history entry by id: %w", err)
	}
	if entry != nil {
		return entry, nil
	}

	entry, err = a.historyStore.GetByAlias(idOrAliasORPosition)
	if err != nil {
		return nil, fmt.Errorf("getting history entry by alias: %w", err)
	}

	return entry, nil
}

func (a *App) getInteractive(params *ConnectToInput) (*historyv1alpha.HistoryEntry, error) {

	entries, err := a.historyStore.GetAllSortedByLastUsed()
	if err != nil {
		return nil, fmt.Errorf("getting history entries: %w", err)
	}
	options, err := a.generateOptions(params, entries)
	if err != nil {
		return nil, fmt.Errorf("getting history entrie options: %w", err)
	}

	selectedEntryString, err := prompt.Choose("history-entry", "Select a history entry", true, prompt.OptionsFromStringSlice(options))
	if err != nil {
		return nil, fmt.Errorf("asking for entry: %w", err)
	}

	// Get the name (ID) of the entry, which is the first alphanumerical string in the row
	nameRegex := regexp.MustCompile("[a-zA-Z0-9]+")
	selectedEntryName := nameRegex.FindString(selectedEntryString)
	if selectedEntryName == "" {
		return nil, history.ErrEntryNotFound
	}
	entry, err := a.historyStore.GetByID(selectedEntryName)
	if err != nil {
		return nil, fmt.Errorf("error getting entry with id: %w", err)
	}
	return entry, nil
}

func (a *App) generateOptions(params *ConnectToInput, entries *historyv1alpha.HistoryEntryList) ([]string, error) {

	options := []string{}
	// Make the history entries table, same output  as the kconnect ls command
	currentContexID, _ := a.getCurrentContextID(params.Kubeconfig)
	entriesTable := entries.ToTable(currentContexID)
	objPrinter, err := printer.New("table")
	if err != nil {
		return nil, fmt.Errorf("making printer: %w", err)
	}
	// Do not print the table to stdout, instead pass it to a byte buffer. We can then convert this to a string and use each row as an option
	buf := new(bytes.Buffer)
	objPrinter.Print(entriesTable, buf)
	tableString := buf.String()
	for i, s := range strings.Split(tableString, "\n") {
		// Ignore the first option as this is the headers. Also ignore empty values
		if i == 0 || s == "" {
			continue
		}
		options = append(options, s)
	}
	return options, nil
}

func (a *App) buildConnectToConfig(configFile string, discoveryProvider string, idProvider string, historyEntry *historyv1alpha.HistoryEntry) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()

	idProviderReg, err := registry.GetIdentityProviderRegistration(idProvider)
	if err != nil {
		return nil, fmt.Errorf("getting identity provider registration: %w", err)
	}
	idCfg, err := idProviderReg.ConfigurationItemsFunc(discoveryProvider)
	if err != nil {
		return nil, fmt.Errorf("getting identity provider config: %w", err)
	}
	discoProviderReg, err := registry.GetDiscoveryProviderRegistration(discoveryProvider)
	if err != nil {
		return nil, fmt.Errorf("getting discovery provider registration: %w", err)
	}
	discoCfg, err := discoProviderReg.ConfigurationItemsFunc(idProvider)
	if err != nil {
		return nil, fmt.Errorf("getting discovery provider config: %w", err)
	}

	if err := cs.AddSet(idCfg); err != nil {
		return nil, fmt.Errorf("adding identity provider config items: %w", err)
	}
	if err := cs.AddSet(discoCfg); err != nil {
		return nil, fmt.Errorf("adding cluster provider config items: %w", err)
	}
	if err := common.AddCommonIdentityConfig(cs); err != nil {
		return nil, fmt.Errorf("adding common identity config items: %w", err)
	}
	if err := common.AddCommonClusterConfig(cs); err != nil {
		return nil, fmt.Errorf("adding common cluster config items: %w", err)
	}
	if err := AddKubeconfigConfigItems(cs); err != nil {
		return nil, fmt.Errorf("adding kubeconfig config items: %w", err)
	}
	if err := AddCommonConfigItems(cs); err != nil {
		return nil, fmt.Errorf("adding common config items: %w", err)
	}
	if err := AddCommonUseConfigItems(cs); err != nil {
		return nil, fmt.Errorf("adding common use config items: %w", err)
	}

	for k, v := range historyEntry.Spec.Flags {
		configItem := cs.Get(k)
		if configItem == nil {
			zap.S().Debugw("no config item found", "name", k)
			continue
		}

		switch configItem.Type {
		case config.ItemTypeString:
			configItem.Value = v
		case config.ItemTypeInt:
			intVal, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			configItem.Value = intVal
		case config.ItemTypeBool:
			boolVal, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			configItem.Value = boolVal
		default:
			return nil, fmt.Errorf("trying to set config item %s of type %s: %w", configItem.Name, configItem.Type, ErrUnknownConfigItemType)
		}
	}

	if err := config.ApplyToConfigSetWithProvider(configFile, cs, discoveryProvider); err != nil {
		return nil, fmt.Errorf("applying app config: %w", err)
	}

	for _, configItem := range cs.GetAll() {
		if !configItem.HasValue() {
			configItem.Value = configItem.DefaultValue
		}
	}

	return cs, nil
}
