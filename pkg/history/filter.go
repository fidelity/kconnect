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

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
)

type FilterSpec struct {
	ClusterProvider  *string
	IdentityProvider *string

	ProviderID *string
	HistoryID  *string
	Alias      *string

	Flags map[string]string

	Kubeconfig *string
}

var (
	DefaultFilterFuncs = []FilterFunc{ByHistoryID, ByProviderID, ByAlias, ByClusterProvider, ByIdentityProvider, ByFlags}

	ErrListNil       = errors.New("history list is nil")
	ErrFilterSpecNil = errors.New("filter spec is nil")
)

// FilterFunc is the spec for a filter function that returns true if
// the history entry should be included. A filterFunc decides how to
// handle values in the spec and what is deemed as empty or not specified
type FilterFunc func(spec *FilterSpec, entry *historyv1alpha.HistoryEntry) bool

func FilterHistory(list *historyv1alpha.HistoryEntryList, filterSpec *FilterSpec) ([]*historyv1alpha.HistoryEntry, error) {
	return FilterHistoryWithFuncs(list, filterSpec, DefaultFilterFuncs)
}

func FilterHistoryWithFuncs(list *historyv1alpha.HistoryEntryList, filterSpec *FilterSpec, filterFucs []FilterFunc) ([]*historyv1alpha.HistoryEntry, error) {
	if list == nil {
		return nil, ErrListNil
	}
	if filterSpec == nil {
		return nil, ErrFilterSpecNil
	}

	entries := []*historyv1alpha.HistoryEntry{}
	for _, entry := range list.Items {
		entryToFilter := &entry
		if FilterEntry(entryToFilter, filterSpec, filterFucs) {
			entries = append(entries, entryToFilter)
		}
	}

	return entries, nil
}

func FilterEntry(entry *historyv1alpha.HistoryEntry, filterSpec *FilterSpec, filterFucs []FilterFunc) bool {
	for _, filterFn := range filterFucs {
		if !filterFn(filterSpec, entry) {
			return false
		}
	}

	return true
}

func ByHistoryID(spec *FilterSpec, entry *historyv1alpha.HistoryEntry) bool {
	if spec.HistoryID == nil || *spec.HistoryID == "" {
		return true
	}

	return entry.Spec.ID == *spec.HistoryID
}

func ByProviderID(spec *FilterSpec, entry *historyv1alpha.HistoryEntry) bool {
	if spec.ProviderID == nil || *spec.ProviderID == "" {
		return true
	}

	return entry.Spec.ProviderID == *spec.ProviderID
}

func ByAlias(spec *FilterSpec, entry *historyv1alpha.HistoryEntry) bool {
	if spec.Alias == nil || *spec.Alias == "" {
		return true
	}

	return *entry.Spec.Alias == *spec.Alias
}

func ByClusterProvider(spec *FilterSpec, entry *historyv1alpha.HistoryEntry) bool {
	if spec.ClusterProvider == nil || *spec.ClusterProvider == "" {
		return true
	}
	return entry.Spec.Provider == *spec.ClusterProvider
}

func ByIdentityProvider(spec *FilterSpec, entry *historyv1alpha.HistoryEntry) bool {
	if spec.IdentityProvider == nil || *spec.IdentityProvider == "" {
		return true
	}
	return entry.Spec.Identity == *spec.IdentityProvider
}

func ByFlags(spec *FilterSpec, entry *historyv1alpha.HistoryEntry) bool {
	if len(spec.Flags) == 0 {
		return true
	}

	return entryHasFlags(entry, spec.Flags)
}

func entryHasFlags(entry *historyv1alpha.HistoryEntry, flags map[string]string) bool {
	for flagKey, flagValue := range flags {
		entryValue, ok := entry.Spec.Flags[flagKey]
		if !ok {
			return false
		}
		if entryValue != flagValue {
			return false
		}
	}

	return true
}
