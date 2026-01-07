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
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	htime "github.com/fidelity/kconnect/pkg/history/time"
	"github.com/oklog/ulid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

// HistoryEntrySpec represents a history item
type HistoryEntrySpec struct {
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
	// LastModified is the date/time that the entry was last modified
	LastModified metav1.Time `json:"lastModified"`
	// LastUsed is the date/time that the entry was last updated
	LastUsed metav1.Time `json:"lastUsed"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HistoryEntry represents a history entry
type HistoryEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   HistoryEntrySpec   `json:"spec"`
	Status HistoryEntryStatus `json:"status"`
}

// +kubebuilder:object:root=true

// HistoryEntryList is a list of history entries
type HistoryEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []HistoryEntry `json:"items"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HistoryReference is a kubeconfig extension to hold a reference to a history item
type HistoryReference struct {
	metav1.TypeMeta `json:",inline"`

	EntryID string
}

var ErrNoHistoryExtension = errors.New("no kconnext history extension found")

var ignoreFlags = map[string]struct{}{
	"profile": {},
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
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0) //nolint: gosec
	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	created := metav1.Now()

	entry := &HistoryEntry{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "HistoryEntry",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              strings.ToLower(id.String()),
			CreationTimestamp: created,
			Generation:        1,
		},
		Spec: HistoryEntrySpec{},
		Status: HistoryEntryStatus{
			LastModified: created,
			LastUsed:     created,
		},
	}

	return entry
}

func NewHistoryReference(entryID string) *HistoryReference {
	return &HistoryReference{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "HistoryReference",
		},
		EntryID: entryID,
	}
}

func GetHistoryReferenceFromContext(context *api.Context) (*HistoryReference, error) {
	kconnectExtension, ok := context.Extensions["kconnect"]
	if !ok {
		return nil, ErrNoHistoryExtension
	}

	b, err := json.Marshal(kconnectExtension)
	if err != nil {
		return nil, fmt.Errorf("marshalling json: %w", err)
	}

	kconnectExtensionObj := HistoryReference{}

	err = json.Unmarshal(b, &kconnectExtensionObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	return &kconnectExtensionObj, nil
}

func (h *HistoryEntry) Equals(other *HistoryEntry) bool {
	if h == nil || other == nil {
		return h == other
	}

	// Compare specific fields
	equals1 := h.Spec.Provider == other.Spec.Provider &&
		h.Spec.Identity == other.Spec.Identity &&
		h.Spec.ProviderID == other.Spec.ProviderID &&
		h.Spec.ConfigFile == other.Spec.ConfigFile
	if !equals1 {
		return false
	}

	return reflect.DeepEqual(filterFlags(h.Spec.Flags), filterFlags(other.Spec.Flags))
}

// filter will create a new map based on flags, without keys that are specifically ignored, and without blank ("") values
func filterFlags(m map[string]string) map[string]string {
	filtered := make(map[string]string)

	for k, v := range m {
		if _, ignore := ignoreFlags[k]; !ignore && v != "" {
			filtered[k] = v
		}
	}

	return filtered
}

func (l *HistoryEntryList) ToTable(currentContextID string) *metav1.Table {
	table := &metav1.Table{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metav1.SchemeGroupVersion.String(),
			Kind:       "Table",
		},
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{
				Name: "Cur",
				Type: "string",
			},
			{
				Name: "Id",
				Type: "string",
			},
			{
				Name: "Alias",
				Type: "string",
			},
			{
				Name: "Provider",
				Type: "string",
			},
			{
				Name: "ProviderID",
				Type: "string",
			},
			{
				Name: "Identity",
				Type: "string",
			},
			{
				Name: "User",
				Type: "string",
			},
			{
				Name: "Time left",
				Type: "String",
			},
		},
	}

	for i, entry := range l.Items {
		var row metav1.TableRow

		currentContextIndicator := ""
		if entry.Name == currentContextID {
			currentContextIndicator = ">"
		}

		timeLeft := getTimeLeft(&l.Items[i])
		username := entry.Spec.Flags["username"]

		row = metav1.TableRow{
			Cells: []any{
				currentContextIndicator,
				entry.ObjectMeta.Name,
				*entry.Spec.Alias,
				entry.Spec.Provider,
				entry.Spec.ProviderID,
				entry.Spec.Identity,
				username,
				timeLeft},
		}

		table.Rows = append(table.Rows, row)
	}

	return table
}

func getTimeLeft(entry *HistoryEntry) string {
	var expiresTime time.Time

	var err error

	if entry.Spec.Provider == "eks" && entry.Spec.Identity == "saml" {
		expiresTime, err = htime.GetExpireTimeFromAWSCredentials(entry.Spec.Flags["aws-profile"])
		if err != nil {
			return ""
		}
	} else {
		return "NA"
	}
	// TODO - other variations e.g. AKS
	return htime.GetRemainingTime(expiresTime)
}
