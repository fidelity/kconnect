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

package v1alpha1

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/oklog/ulid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HistoryEntrySpec represents a history item
type HistoryEntrySpec struct {
	// ID is the unique identifier for this history item
	ID string `json:"id"`
	// Provider is the name of the discovery provider
	Provider string `json:"provider"`
	// Identity is the name of the identity provider
	Identity string `json:"identity"`
	// ProviderID is the unique identifier for cluster with the provider
	ProviderID string `json:"providerID"`
	// Flags is the non sensitive flags and values
	Flags map[string]string `json:"flags,omitempty"`
	// ConfigFile is the path to the config file that was updated
	ConfigFile string `json:"configFile"`
	// Alias is the given alternative user friendly name for the connection
	Alias *string `json:"alias,omitempty"`
}

type HistoryEntryStatus struct {
	// LastUpdated is the date/time that the entry was last updated
	LastUpdated metav1.Time `json:"lastUpdated"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HistoryEntry represents a history entry
type HistoryEntry struct {
	metav1.TypeMeta `json:",inline"`

	Spec   HistoryEntrySpec   `json:"spec,omitempty"`
	Status HistoryEntryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HistoryEntryList is a list of history entries
type HistoryEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HistoryEntry `json:"items"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HistoryExtension is a kubeconfig extension to hold a reference to a history item
type HistoryExtension struct {
	metav1.TypeMeta `json:",inline"`

	HistoryEntryID string
}

func NewHistoryEntryList() *HistoryEntryList {
	return &HistoryEntryList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "HistoryEntryList",
		},
		Items: []HistoryEntry{},
	}
}

func NewHistoryEntry() *HistoryEntry {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	entry := &HistoryEntry{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "HistoryEntry",
		},
		Spec: HistoryEntrySpec{
			ID: id.String(),
		},
		Status: HistoryEntryStatus{
			LastUpdated: metav1.Now(),
		},
	}

	return entry
}

func NewHistoryExtension(historyID string) *HistoryExtension {
	return &HistoryExtension{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "HistoryExtension",
		},
		HistoryEntryID: historyID,
	}
}

func (h *HistoryEntry) Equals(other *HistoryEntry) bool {
	if h == nil || other == nil {
		return h == other
	}

	if h == other {
		return true
	}

	// TODO: we could do explicit comparision of the fields

	return reflect.DeepEqual(h.Spec, other.Spec)
}
