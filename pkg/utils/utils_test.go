package utils

import "testing"

func Test_Filter(t *testing.T) {
	testCases := []struct {
		name        string
		inputValue  string
		inputFilter string
		expect      bool
	}{
		{
			name:        "Simple match",
			inputValue:  "this is a simple string",
			inputFilter: "simple",
			expect:      true,
		},
		{
			name:        "Simple mismatch",
			inputValue:  "this is a simple string",
			inputFilter: "isnt",
			expect:      false,
		},
		{
			name:        "Multiple match (*)",
			inputValue:  "this is a simple string",
			inputFilter: "simple*this",
			expect:      true,
		},
		{
			name:        "Multiple mismatch (*)",
			inputValue:  "this is a simple string",
			inputFilter: "string*isnt",
			expect:      false,
		},
		{
			name:        "Multiple match (whitespace)",
			inputValue:  "this is a simple string",
			inputFilter: "string this",
			expect:      true,
		},
		{
			name:        "Multiple mismatch (whitespace)",
			inputValue:  "this is a simple string",
			inputFilter: "string mismatch",
			expect:      false,
		},
		{
			name:        "Multiple match (* + whitespace)",
			inputValue:  "this is a simple string",
			inputFilter: "string*this*string",
			expect:      true,
		},
		{
			name:        "Multiple mismatch (* + whitespace)",
			inputValue:  "this is a simple string",
			inputFilter: "string*is a mismatch",
			expect:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := SurveyFilter(tc.inputFilter, tc.inputValue, 0)
			if actual != tc.expect {
				t.Fatalf("expected %t but got %t", tc.expect, actual)
			}
		})
	}

}
