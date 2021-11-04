package utils

import (
	"reflect"
	"testing"
)

func Test_SurveyFilter(t *testing.T) {
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



func Test_RegexFilter(t *testing.T) {
	testCases := []struct {
		name                    string
		inputRegex              string
		inputOptions            []string
		expectedFilteredOptions []string
		expectErr               bool
	}{
		{
			name:                    "empty filter",
			inputRegex:              "",
			inputOptions:            []string{"test1", "test2"},
			expectedFilteredOptions: []string{"test1", "test2"},
			expectErr:               false,
		},
		{
			name:                    "simple filter",
			inputRegex:              "test",
			inputOptions:            []string{"test1", "test2", "blah"},
			expectedFilteredOptions: []string{"test1", "test2"},
			expectErr:               false,
		},
		{
			name:                    "exact match filter",
			inputRegex:              "^test2$",
			inputOptions:            []string{"1test2", "test21", "test2"},
			expectedFilteredOptions: []string{"test2"},
			expectErr:               false,
		},
		{
			name:                    "multiple filter",
			inputRegex:              "^us-east|^eu-west-2$",
			inputOptions:            []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "af-south-1", "ap-east-1", "ap-south-1", "ap-northeast-3", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ca-central-1", "cn-north-1", "cn-northwest-1", "eu-central-1", "eu-west-1", "eu-west-2", "eu-south-1", "eu-west-3", "eu-north-1", "me-south-1", "sa-east-1"},
			expectedFilteredOptions: []string{"us-east-1", "us-east-2", "eu-west-2"},
			expectErr:               false,
		},
		{
			name:                    "bad filter",
			inputRegex:              "*",
			inputOptions:            []string{"test1", "test2"},
			expectedFilteredOptions: []string{},
			expectErr:               true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualFilteredOptions, actualErr := RegexFilter(tc.inputOptions, tc.inputRegex)
			if actualErr != nil {
				if !tc.expectErr {
					t.Fatalf("expected %t but got %t", tc.expectErr, actualErr)
				}	
			} else {
				if !reflect.DeepEqual(tc.expectedFilteredOptions, actualFilteredOptions) {
					t.Fatalf("expected %v but got %v", tc.expectedFilteredOptions, actualFilteredOptions)
				}
			}
		})
	}

}


