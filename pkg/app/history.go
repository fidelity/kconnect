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
	"maps"

	"github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"go.uber.org/zap"
)

type HistoryImportInput struct {
	CommonConfig
	HistoryLocationConfig
	HistoryImportConfig
}

type HistoryExportInput struct {
	CommonConfig
	HistoryLocationConfig
	HistoryExportConfig
}

type HistoryRemoveInput struct {
	CommonConfig
	HistoryLocationConfig
	HistoryRemoveConfig

	RemoveList []string
}

// HistoryImport implements the history import functionality
func (a *App) HistoryImport(ctx context.Context, input *HistoryImportInput) error {
	zap.S().Infow("importing history")

	filterSpec := createFilter(input.Filter)
	setFlags := flags.ParseFlagMultiValueToMap(input.Set)

	importList, err := readImportFile(input.File)
	if err != nil {
		return fmt.Errorf("reading history import file: %w", err)
	}

	err = history.FilterHistory(importList, filterSpec)
	if err != nil {
		return fmt.Errorf("filtering history list: %w", err)
	}

	var historyList = &v1alpha1.HistoryEntryList{}
	if !input.Clean {
		historyList, err = a.historyStore.GetAll()
		if err != nil {
			return fmt.Errorf("getting existing history list: %w", err)
		}
	} else {
		zap.S().Info("deleting exiting history")
	}

	importCount := 0

	for i := range importList.Items {
		newEntry := processEntry(&importList.Items[i], setFlags)

		inUse, index := checkAliasInUse(historyList, newEntry)
		if inUse {
			if input.Overwrite {
				zap.S().Infow("Entry with alias already exists, overwriting", "alias", newEntry.Spec.Alias)
				historyList.Items[index] = *newEntry
				importCount++
			} else {
				zap.S().Infow("Entry with alias already exists, skipping", "alias", newEntry.Spec.Alias)
			}
		} else {
			historyList.Items = append(historyList.Items, *newEntry)
			importCount++
		}
	}

	zap.S().Infof("Importing %d entries", importCount)

	err = a.historyStore.SetHistoryList(historyList)
	if err != nil {
		return fmt.Errorf("storing history entries: %w", err)
	}

	return nil
}

// HistoryExport implements the history export functionality
func (a *App) HistoryExport(ctx context.Context, input *HistoryExportInput) error {
	zap.S().Infow("exporting history")

	filterSpec := createFilter(input.Filter)
	setFlags := flags.ParseFlagMultiValueToMap(input.Set)

	historyList, err := a.historyStore.GetAll()
	if err != nil {
		return fmt.Errorf("getting history list: %w", err)
	}

	err = history.FilterHistory(historyList, filterSpec)
	if err != nil {
		return fmt.Errorf("filtering history list: %w", err)
	}

	var historyExportList = &v1alpha1.HistoryEntryList{}

	exportCount := 0

	for i := range historyList.Items {
		newEntry := processEntry(&historyList.Items[i], setFlags)
		historyExportList.Items = append(historyExportList.Items, *newEntry)
		exportCount++
	}

	zap.S().Infof("exporting %d entries", exportCount)

	err = writeExportFile(input.File, historyExportList)
	if err != nil {
		return fmt.Errorf("writing export file: %w", err)
	}

	return nil
}

func (a *App) HistoryRemove(ctx context.Context, input *HistoryRemoveInput) error {
	zap.S().Infow("removing history")

	historyList, err := a.historyStore.GetAll()
	if err != nil {
		return fmt.Errorf("getting history list: %w", err)
	}

	var entriesToRemove []*v1alpha1.HistoryEntry

	switch {
	case input.All:
		for i := range historyList.Items {
			entriesToRemove = append(entriesToRemove, &historyList.Items[i])
		}
	case input.Filter != "":
		filterSpec := createFilter(input.Filter)

		err = history.FilterHistory(historyList, filterSpec)
		if err != nil {
			return fmt.Errorf("filtering history list: %w", err)
		}

		for i := range historyList.Items {
			entriesToRemove = append(entriesToRemove, &historyList.Items[i])
		}
	default:
		for _, entryID := range input.RemoveList {
			entry, err := a.historyStore.GetByID(entryID)
			if err != nil {
				return fmt.Errorf("getting history entry: %w", err)
			}

			entriesToRemove = append(entriesToRemove, entry)
		}
	}

	zap.S().Infof("removing %d entries", len(entriesToRemove))

	err = a.historyStore.Remove(entriesToRemove)
	if err != nil {
		return fmt.Errorf("removing history entries: %w", err)
	}

	return nil
}

func createFilter(filterString string) *history.FilterSpec {
	filterParts := flags.ParseFlagMultiValueToMap(filterString)
	return history.CreateFilterFromMap(filterParts)
}

func processEntry(entry *v1alpha1.HistoryEntry, overwriteFlags map[string]string) *v1alpha1.HistoryEntry {
	newEntry := v1alpha1.NewHistoryEntry()

	newEntry.Spec = entry.Spec
	maps.Copy(newEntry.Spec.Flags, overwriteFlags)

	return newEntry
}

func readImportFile(location string) (*v1alpha1.HistoryEntryList, error) {
	fileLoader, err := loader.NewFileLoader(location)
	if err != nil {
		return nil, err
	}

	return fileLoader.Load()
}

func writeExportFile(location string, historyList *v1alpha1.HistoryEntryList) error {
	fileLoader, err := loader.NewFileLoader(location)
	if err != nil {
		return err
	}

	return fileLoader.Save(historyList)
}

func checkAliasInUse(historyList *v1alpha1.HistoryEntryList, entryToCheck *v1alpha1.HistoryEntry) (bool, int) {
	for i, entry := range historyList.Items {
		if *entry.Spec.Alias == *entryToCheck.Spec.Alias {
			return true, i
		}
	}

	return false, -1
}
