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
	"regexp"
	"testing"
)

func TestNewHistoryEntryIDsAreUnique(t *testing.T) {
	// Entries created one after another can share the same clock reading, especially on
	// platforms with a coarse clock, and must still be given different ids.
	const count = 5000

	validID := regexp.MustCompile("^[0-9a-z]{26}$")
	seen := make(map[string]bool, count)

	for range count {
		id := NewHistoryEntry().Name

		if !validID.MatchString(id) {
			t.Fatalf("id %q is not a lower case ULID", id)
		}

		if seen[id] {
			t.Fatalf("id %q was generated more than once", id)
		}

		seen[id] = true
	}
}

func TestGetAlias(t *testing.T) {
	entry := NewHistoryEntry()
	if alias := entry.GetAlias(); alias != "" {
		t.Fatalf("expected no alias for a new entry, got %q", alias)
	}

	alias := "dev"
	entry.Spec.Alias = &alias

	if got := entry.GetAlias(); got != "dev" {
		t.Fatalf("expected dev, got %q", got)
	}

	var none *HistoryEntry
	if got := none.GetAlias(); got != "" {
		t.Fatalf("expected no alias for a nil entry, got %q", got)
	}
}
