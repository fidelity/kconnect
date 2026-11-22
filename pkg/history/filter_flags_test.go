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
	"testing"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
)

func Test_EqualsWithWildcardLiteralCharacters(t *testing.T) {
	testCases := []struct {
		name        string
		filterInput string
		input       string
		expect      bool
	}{
		{name: "unbalanced bracket does not panic", filterInput: "prod(", input: "prod(", expect: true},
		{name: "unbalanced bracket mismatch", filterInput: "prod(", input: "prod", expect: false},
		{name: "square bracket does not panic", filterInput: "[abc", input: "[abc", expect: true},
		{name: "dot is literal", filterInput: "a.c", input: "abc", expect: false},
		{name: "dot matches itself", filterInput: "a.c", input: "a.c", expect: true},
		{name: "plus is literal", filterInput: "a+", input: "aaa", expect: false},
		{name: "question mark is literal", filterInput: "ab?", input: "a", expect: false},
		{name: "arn with wildcard", filterInput: "arn:aws:eks:us-east-1:*:cluster/dev", input: "arn:aws:eks:us-east-1:123456789012:cluster/dev", expect: true},
		{name: "url with wildcard", filterInput: "https://idp.example.com/*", input: "https://idp.example.com/adfs?x=1", expect: true},
		{name: "url is not a pattern", filterInput: "https://idp.example.com/", input: "https://idpXexample.com/", expect: false},
		{name: "several wildcards", filterInput: "*us-*-1*", input: "eu-west-1", expect: false},
		{name: "several wildcards match", filterInput: "*us-*-1*", input: "arn:us-east-1:x", expect: true},
		{name: "only a wildcard", filterInput: "*", input: "", expect: true},
		{name: "empty filter matches only empty", filterInput: "", input: "something", expect: false},
		{name: "backslash is literal", filterInput: `dom\user`, input: `dom\user`, expect: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := equalsWithWildcard(tc.filterInput, tc.input)
			if actual != tc.expect {
				t.Fatalf("expected %t but got %t", tc.expect, actual)
			}
		})
	}
}

func Test_EntryHasFlagsChecksEveryFlag(t *testing.T) {
	entry := historyv1alpha.NewHistoryEntry()
	entry.Spec.Flags = map[string]string{"region": "us-east-1", "namespace": "dev", "role": "admin"}

	testCases := []struct {
		name   string
		flags  map[string]string
		expect bool
	}{
		{name: "all match", flags: map[string]string{"region": "us-east-1", "namespace": "dev"}, expect: true},
		{name: "first matches, second does not", flags: map[string]string{"region": "us-east-1", "namespace": "prod"}, expect: false},
		{name: "none match", flags: map[string]string{"region": "eu-west-1", "namespace": "prod"}, expect: false},
		{name: "three flags, one mismatch", flags: map[string]string{"region": "us-east-1", "namespace": "dev", "role": "reader"}, expect: false},
		{name: "three flags, all match with a wildcard", flags: map[string]string{"region": "us-*", "namespace": "dev", "role": "admin"}, expect: true},
		{name: "flag the entry doesn't have", flags: map[string]string{"region": "us-east-1", "tenant": "x"}, expect: false},
		{name: "no flags", flags: map[string]string{}, expect: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Map iteration order is random, so repeat to cover every order
			for range 200 {
				if actual := entryHasFlags(entry, tc.flags); actual != tc.expect {
					t.Fatalf("expected %t but got %t", tc.expect, actual)
				}
			}
		})
	}
}

func Test_FilterEntriesWithoutAnAlias(t *testing.T) {
	withAlias := historyv1alpha.NewHistoryEntry()
	alias := "dev"
	withAlias.Spec.Alias = &alias

	// For example an entry from a history file that was written by hand
	withoutAlias := historyv1alpha.NewHistoryEntry()
	withoutAlias.Spec.Alias = nil

	list := historyv1alpha.NewHistoryEntryList()
	list.Items = append(list.Items, *withAlias, *withoutAlias)

	filter := "dev"
	if err := FilterHistory(list, &FilterSpec{Alias: &filter}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(list.Items) != 1 || list.Items[0].GetAlias() != "dev" {
		t.Fatalf("expected only the entry with the alias, got %d entries", len(list.Items))
	}
}
