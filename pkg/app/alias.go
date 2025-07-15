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

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/printer"
)

// AliasListInput defines the inputs for AliasList
type AliasListInput struct {
	CommonConfig
	HistoryLocationConfig

	Output *printer.OutputPrinter `json:"output,omitempty"`
}

// AliasAddInput defines the inputs for AliasAdd
type AliasAddInput struct {
	CommonConfig
	HistoryLocationConfig
	HistoryIdentifierConfig
}

// AliasRemoveInput defines the inputs for AliasRemove
type AliasRemoveInput struct {
	CommonConfig
	HistoryLocationConfig
	HistoryIdentifierConfig

	All bool `json:"all"`
}

// AliasList implements the alias listing functionality
func (a *App) AliasList(ctx context.Context, input *AliasListInput) error {
	zap.S().Infow("listing aliases")

	list, err := a.historyStore.GetAll()
	if err != nil {
		return fmt.Errorf("getting history entries: %w", err)
	}

	aliases := []string{}

	for _, entry := range list.Items {
		if entry.Spec.Alias != nil && *entry.Spec.Alias != "" {
			aliases = append(aliases, *entry.Spec.Alias)
		}
	}

	objPrinter, err := printer.New(*input.Output)
	if err != nil {
		return fmt.Errorf("getting printer for output %s: %w", *input.Output, err)
	}

	if *input.Output == printer.OutputPrinterTable {
		table := printer.ConvertSliceToTable("Alias", aliases)
		return objPrinter.Print(table, os.Stdout)
	}

	return objPrinter.Print(aliases, os.Stdout)
}

// AliasAdd will add an alias to an existing history entry
func (a *App) AliasAdd(ctx context.Context, input *AliasAddInput) error {
	zap.S().Infow("adding alias to history entry", "id", input.ID, "alias", input.Alias)

	if input.Alias == "" {
		return ErrAliasRequired
	}

	if input.ID == "" {
		return ErrHistoryIDRequired
	}

	aliasInUse, err := a.aliasInUse(&input.Alias)
	if err != nil {
		return fmt.Errorf("checking if alias in use: %w", err)
	}

	if aliasInUse {
		return ErrAliasAlreadyUsed
	}

	entry, err := a.historyStore.GetByID(input.ID)
	if err != nil {
		return fmt.Errorf("getting history entry: %w", err)
	}

	if entry == nil {
		return history.ErrEntryNotFound
	}

	entry.Spec.Alias = &input.Alias
	entry.Status.LastModified = v1.Now()

	zap.S().Debug("updating history entry with new alias")

	if err := a.historyStore.Update(entry); err != nil {
		return fmt.Errorf("updating history entry with alias: %w", err)
	}

	zap.S().Info("aliases added to histiry entry")

	return nil
}

// AliasRemove will remove an alias from history entries. You can also remove all
// aliases.
func (a *App) AliasRemove(ctx context.Context, input *AliasRemoveInput) error {
	zap.S().Infow("removing aliases from history entries", "id", input.ID, "alias", input.Alias, "all", input.All)

	if input.Alias == "" && input.ID == "" && !input.All {
		zap.S().Warn("no remove criteria specified, no action taken")
		return nil
	}

	if input.Alias != "" && input.ID != "" {
		return ErrAliasAndIDNotAllowed
	}

	if input.All && (input.Alias != "" || input.ID != "") {
		zap.S().Warn("all specified along with an alias or id, any id or alias specified will be ignored")
	}

	found, err := a.getAliasEntries(input.ID, input.Alias, input.All)
	if err != nil {
		return fmt.Errorf("getting alias entries: %w", err)
	}

	if len(found) == 0 {
		zap.S().Info("no history entries found with the matching alias details, no action taken")
		return nil
	}

	for i := range found {
		oldalias := *found[i].Spec.Alias

		updatedEntry := found[i].DeepCopy()

		*updatedEntry.Spec.Alias = ""
		updatedEntry.Status.LastModified = v1.Now()

		zap.S().Debugw("updating history entry to remove alias", "id", updatedEntry.ObjectMeta.Name, "oldalias", oldalias)

		if err := a.historyStore.Update(updatedEntry); err != nil {
			return fmt.Errorf("updating history entry: %w", err)
		}
	}

	zap.S().Infow("removed aliases from history entries", "count", len(found))

	return nil
}

func (a *App) getAliasEntries(id string, alias string, all bool) ([]*apiv1alpha.HistoryEntry, error) {
	var found []*apiv1alpha.HistoryEntry

	if all { //nolint: nestif
		list, err := a.historyStore.GetAll()
		if err != nil {
			return nil, fmt.Errorf("getting history entries: %w", err)
		}

		for _, entry := range list.Items {
			if entry.Spec.Alias != nil && *entry.Spec.Alias != "" {
				entryToUpdate := entry.DeepCopy()
				found = append(found, entryToUpdate)
			}
		}
	} else {
		if alias != "" {
			entry, err := a.historyStore.GetByAlias(alias)
			if err != nil {
				return nil, fmt.Errorf("getting history entry by alias: %w", err)
			}

			if entry != nil {
				found = append(found, entry)
			}
		}

		if id != "" {
			entry, err := a.historyStore.GetByID(id)
			if err != nil {
				return nil, fmt.Errorf("getting history entry by id: %w", err)
			}

			if entry != nil {
				found = append(found, entry)
			}
		}
	}

	return found, nil
}
