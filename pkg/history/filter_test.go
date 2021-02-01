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
)

func Test_EqualsWithWildcard(t *testing.T) {
	testCases := []struct {
		name            string
		filterInput     string
		input           string
		expect          bool
	}{
		{
			name:            "no wildcards match",
			filterInput:     "exactmatch",
		    input:           "exactmatch",
			expect:          true,
		},
		{
			name:            "no wildcards mistmatch",
			filterInput:     "exactmatch",
		    input:           "nomatch",
			expect:          false,
		},
		{
			name:            "wildcard match start",
			filterInput:     "*match",
		    input:           "thisshouldmatch",
			expect:          true,
		},
		{
			name:            "wildcard mismatch start",
			filterInput:     "*match",
		    input:           "thisshouldmatchNOT",
			expect:          false,
		},
		{
			name:            "wildcard match end",
			filterInput:     "match*",
		    input:           "matchthisshould",
			expect:          true,
		},
		{
			name:            "wildcard mismatch end",
			filterInput:     "match*",
		    input:           "thishouldnotmatch",
			expect:          false,
		},
		{
			name:            "wildcard match start and end",
			filterInput:     "*match*",
		    input:           "thisshouldmatchitshould",
			expect:          true,
		},
		{
			name:            "wildcard mismatch start and end",
			filterInput:     "*match*",
		    input:           "thisshouldnotmaaatchno",
			expect:          false,
		},
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