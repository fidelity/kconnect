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

package flags

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

func TestExistsWithValue(t *testing.T) {
	testCases := []struct {
		name    string
		flag    string
		flagset *pflag.FlagSet
		expect  bool
	}{
		{
			name:    "flag doesn't exists in flagset",
			flag:    "flag1",
			flagset: pflag.NewFlagSet("", pflag.PanicOnError),
			expect:  false,
		},
		{
			name:    "flag exists but value is empty string",
			flag:    "flag1",
			flagset: createTestFlagSet(t, "flag1", ""),
			expect:  false,
		},
		{
			name:    "flag exists and value is set",
			flag:    "flag1",
			flagset: createTestFlagSet(t, "flag1", "this is set"),
			expect:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			actual := ExistsWithValue(tc.flag, tc.flagset)
			g.Expect(actual).To(Equal(tc.expect))
		})
	}
}

func createTestFlagSet(t *testing.T, name, value string) *pflag.FlagSet {
	fs := pflag.NewFlagSet("", pflag.PanicOnError)
	fs.String(name, "", "test flag")

	flagName := fmt.Sprintf("--%s", name)

	if err := fs.Parse([]string{flagName, value}); err != nil {
		t.Fatal("error parsing flags")
	}

	// Add some other flags
	fs.Bool("bflag", true, "some bool flag")
	fs.String("sflaf", "hello", "some string flag")

	return fs
}
