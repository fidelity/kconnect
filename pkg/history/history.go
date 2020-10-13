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
	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
)

// Store is an interface for the history store
type Store interface {
	Add(entry *historyv1alpha.HistoryEntry) error
	Remove(entries []*historyv1alpha.HistoryEntry) error

	GetAll() (*historyv1alpha.HistoryEntryList, error)
	GetByID(id string) (*historyv1alpha.HistoryEntry, error)
	GetByProvider(providerName string) ([]*historyv1alpha.HistoryEntry, error)
	GetByProviderWithID(providerName, providerID string) ([]*historyv1alpha.HistoryEntry, error)
	GetByAlias(alias string) (*historyv1alpha.HistoryEntry, error)
	GetLastModified() (*historyv1alpha.HistoryEntry, error)

	Update(entry *historyv1alpha.HistoryEntry) error
}
