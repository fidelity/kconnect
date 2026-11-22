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

package prompt_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/fidelity/kconnect/pkg/defaults"
	"github.com/fidelity/kconnect/pkg/prompt"
)

func TestPageSize(t *testing.T) {
	testCases := []struct {
		name           string
		numOptions     int
		terminalHeight int
		expect         int
	}{
		{name: "unknown terminal height uses the default", numOptions: 100, terminalHeight: 0, expect: defaults.DefaultUIPageSize},
		{name: "short terminal still shows the default", numOptions: 100, terminalHeight: 12, expect: defaults.DefaultUIPageSize},
		{name: "standard 24 line terminal", numOptions: 100, terminalHeight: 24, expect: 18},
		{name: "tall terminal is capped", numOptions: 100, terminalHeight: 80, expect: 30},
		{name: "never more than the number of options", numOptions: 4, terminalHeight: 50, expect: 4},
		{name: "fewer options than the default", numOptions: 3, terminalHeight: 0, expect: 3},
		{name: "exactly the default number of options", numOptions: 10, terminalHeight: 0, expect: 10},
		{name: "no options", numOptions: 0, terminalHeight: 24, expect: 18},
		{name: "negative terminal height", numOptions: 100, terminalHeight: -1, expect: defaults.DefaultUIPageSize},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			size := prompt.PageSize(tc.numOptions, tc.terminalHeight)
			g.Expect(size).To(Equal(tc.expect))
			g.Expect(size).To(BeNumerically(">", 0))
		})
	}
}

func TestChooseSelectsTheOnlyOption(t *testing.T) {
	g := NewWithT(t)

	// With a single option there is nothing to ask, so no terminal is needed
	value, err := prompt.Choose("item", "Select an item", true, prompt.OptionsFromMap(map[string]string{"Display name": "the-value"}))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(value).To(Equal("the-value"))
}

func TestChooseInOrderSelectsTheOnlyOption(t *testing.T) {
	g := NewWithT(t)

	value, err := prompt.ChooseInOrder("item", "Select an item", true, []string{"only"})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(value).To(Equal("only"))
}
