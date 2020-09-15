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

package history

import (
	"errors"
	"fmt"
	"reflect"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/history/loader"
)

var (
	ErrNoLoader          = errors.New("loader required for history store")
	ErrStoreFileRequired = errors.New("history store must be a file")
	ErrEntryNotFound     = errors.New("history entry not found")
	ErrDuplicateAlias    = errors.New("duplicate alias detected")
)

func NewStore(maxHistoryItems int, loader loader.Loader) (Store, error) {
	if loader == nil {
		return nil, ErrNoLoader
	}

	return &storeImpl{
		maxHistory: maxHistoryItems,
		loader:     loader,
	}, nil
}

type storeImpl struct {
	loader     loader.Loader
	maxHistory int
}

func (s *storeImpl) Add(entry *historyv1alpha.HistoryEntry) error {
	historyList, err := s.loader.Load()
	if err != nil {
		return fmt.Errorf("reading history file: %w", err)
	}

	historyList.Items = append(historyList.Items, *entry)

	if len(historyList.Items) > s.maxHistory {
		if err := s.trimHistory(historyList); err != nil {
			return fmt.Errorf("trimming history: %w", err)
		}
	}

	return s.loader.Save(historyList)
}

func (s *storeImpl) Remove(entries []*historyv1alpha.HistoryEntry) error {
	historyList, err := s.loader.Load()
	if err != nil {
		return fmt.Errorf("reading history file: %w", err)
	}

	for _, entryToRemove := range entries {
		if err := s.removeEntryFromHistory(historyList, entryToRemove); err != nil {
			return fmt.Errorf("error removing history item %s: %w", entryToRemove.Spec.ID, err)
		}
	}

	return s.loader.Save(historyList)
}

func (s *storeImpl) GetAll() (*historyv1alpha.HistoryEntryList, error) {
	historyList, err := s.loader.Load()
	if err != nil {
		return nil, fmt.Errorf("reading history file: %w", err)
	}

	return historyList, nil
}

func (s *storeImpl) GetByID(id string) (*historyv1alpha.HistoryEntry, error) {
	entries, err := s.filterHistory(func(entry *historyv1alpha.HistoryEntry) bool {
		return entry.Spec.ID == id
	})
	if err != nil {
		return nil, fmt.Errorf("filtering history to id %s: %w", id, err)
	}

	if len(entries) == 0 {
		return nil, nil
	}

	return entries[0], nil
}

func (s *storeImpl) GetByProvider(providerName string) ([]*historyv1alpha.HistoryEntry, error) {
	entries, err := s.filterHistory(func(entry *historyv1alpha.HistoryEntry) bool {
		return entry.Spec.ProviderID == providerName
	})
	if err != nil {
		return nil, fmt.Errorf("filtering history by provider %s: %w", providerName, err)
	}

	return entries, nil
}

func (s *storeImpl) GetByProviderWithID(providerName, providerID string) ([]*historyv1alpha.HistoryEntry, error) {
	entries, err := s.filterHistory(func(entry *historyv1alpha.HistoryEntry) bool {
		return entry.Spec.ProviderID == providerName && entry.Spec.ID == providerID
	})
	if err != nil {
		return nil, fmt.Errorf("filtering history by provider %s and id %s: %w", providerName, providerID, err)
	}

	return entries, nil
}

func (s *storeImpl) GetByAlias(alias string) (*historyv1alpha.HistoryEntry, error) {
	entries, err := s.filterHistory(func(entry *historyv1alpha.HistoryEntry) bool {
		return entry.Spec.Alias != nil && entry.Spec.Alias == &alias
	})
	if err != nil {
		return nil, fmt.Errorf("filtering history by alias %s: %w", alias, err)
	}

	if len(entries) > 1 {
		return nil, ErrDuplicateAlias
	}
	if len(entries) == 0 {
		return nil, nil
	}

	return entries[0], nil
}

func (s *storeImpl) trimHistory(historyList *historyv1alpha.HistoryEntryList) error {
	diff := len(historyList.Items) - s.maxHistory
	if diff < 1 {
		return nil
	}

	historyList.Items = historyList.Items[diff:]

	return nil
}

func (s *storeImpl) removeEntryFromHistory(historyList *historyv1alpha.HistoryEntryList, entryToRemove *historyv1alpha.HistoryEntry) error {
	for i := range historyList.Items {
		entry := historyList.Items[i]

		if reflect.DeepEqual(entry, *entryToRemove) {
			historyList.Items = append(historyList.Items[:i], historyList.Items[i+1:]...)
			return nil
		}
	}

	return ErrEntryNotFound
}

func (s *storeImpl) filterHistory(filter func(entry *historyv1alpha.HistoryEntry) bool) ([]*historyv1alpha.HistoryEntry, error) {
	historyList, err := s.loader.Load()
	if err != nil {
		return nil, fmt.Errorf("reading history file: %w", err)
	}

	filteredEntries := []*historyv1alpha.HistoryEntry{}
	for _, entry := range historyList.Items {
		if filter(&entry) {
			filteredEntries = append(filteredEntries, &entry)
		}
	}

	return filteredEntries, nil
}
