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

package flags_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/fidelity/kconnect/pkg/flags"
)

func TestGetFlagValueFromArgs(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		longName    string
		shortName   string
		expectValue string
		expectFound bool
	}{
		{
			name:     "no args",
			args:     []string{},
			longName: "verbosity", shortName: "v",
		},
		{
			name:     "flag not present",
			args:     []string{"kconnect", "use", "eks", "--region", "us-east-1"},
			longName: "verbosity", shortName: "v",
		},
		{
			name:     "long form with a space",
			args:     []string{"kconnect", "use", "eks", "--idp-protocol", "saml"},
			longName: "idp-protocol", expectValue: "saml", expectFound: true,
		},
		{
			name:     "long form with an equals",
			args:     []string{"kconnect", "use", "eks", "--idp-protocol=saml"},
			longName: "idp-protocol", expectValue: "saml", expectFound: true,
		},
		{
			name:     "short form with a space",
			args:     []string{"kconnect", "-v", "5", "ls"},
			longName: "verbosity", shortName: "v", expectValue: "5", expectFound: true,
		},
		{
			name:     "short form with an equals",
			args:     []string{"kconnect", "-v=5", "ls"},
			longName: "verbosity", shortName: "v", expectValue: "5", expectFound: true,
		},
		{
			name:     "long form with an equals and a short name",
			args:     []string{"kconnect", "--verbosity=9", "ls"},
			longName: "verbosity", shortName: "v", expectValue: "9", expectFound: true,
		},
		{
			name:     "value containing an equals",
			args:     []string{"kconnect", "use", "eks", "--idp-endpoint=https://idp.example.com/?a=b"},
			longName: "idp-endpoint", expectValue: "https://idp.example.com/?a=b", expectFound: true,
		},
		{
			name:     "empty value with an equals",
			args:     []string{"kconnect", "use", "eks", "--idp-protocol="},
			longName: "idp-protocol", expectValue: "", expectFound: true,
		},
		{
			name:     "long flag is the last argument",
			args:     []string{"kconnect", "use", "eks", "--idp-protocol"},
			longName: "idp-protocol",
		},
		{
			name:     "short flag is the last argument",
			args:     []string{"kconnect", "ls", "-v"},
			longName: "verbosity", shortName: "v",
		},
		{
			name:     "another flag with the same prefix is not matched",
			args:     []string{"kconnect", "use", "eks", "--idp-protocol-extra", "x", "--idp-protocols=y"},
			longName: "idp-protocol",
		},
		{
			name:     "short name is not matched when there is none",
			args:     []string{"kconnect", "-", "value"},
			longName: "verbosity",
		},
		{
			name:     "first occurrence is used",
			args:     []string{"kconnect", "--idp-protocol", "saml", "--idp-protocol=aws-iam"},
			longName: "idp-protocol", expectValue: "saml", expectFound: true,
		},
		{
			name:     "flags after a double dash are positional",
			args:     []string{"kconnect", "to", "--", "--idp-protocol", "saml"},
			longName: "idp-protocol",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			value, found := flags.GetFlagValueFromArgs(tc.args, tc.longName, tc.shortName)
			g.Expect(found).To(Equal(tc.expectFound))
			g.Expect(value).To(Equal(tc.expectValue))
		})
	}
}

func TestGetFlagValueDirectFromArgs(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		expect string
	}{
		{name: "space form", args: []string{"kconnect", "-v", "5"}, expect: "5"},
		{name: "equals form", args: []string{"kconnect", "--verbosity=7"}, expect: "7"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			value, err := flags.GetFlagValueDirect(tc.args, "verbosity", "v")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(value).To(Equal(tc.expect))
		})
	}
}

func TestParseFlagMultiValueToMapEdgeCases(t *testing.T) {
	testCases := []struct {
		name   string
		flag   string
		expect map[string]string
	}{
		{
			name:   "empty",
			flag:   "",
			expect: map[string]string{},
		},
		{
			name:   "item without a value does not panic and is ignored",
			flag:   "alias",
			expect: map[string]string{},
		},
		{
			name:   "item without a value amongst valid items",
			flag:   "region=us-east-1,alias,provider=eks",
			expect: map[string]string{"region": "us-east-1", "provider": "eks"},
		},
		{
			name:   "value containing an equals is kept whole",
			flag:   "idp-endpoint=https://idp.example.com/?a=b",
			expect: map[string]string{"idp-endpoint": "https://idp.example.com/?a=b"},
		},
		{
			name:   "empty value is kept",
			flag:   "alias=",
			expect: map[string]string{"alias": ""},
		},
		{
			name:   "item without a key is ignored",
			flag:   "=value,alias=dev",
			expect: map[string]string{"alias": "dev"},
		},
		{
			name:   "trailing comma",
			flag:   "alias=dev,",
			expect: map[string]string{"alias": "dev"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(flags.ParseFlagMultiValueToMap(tc.flag)).To(Equal(tc.expect))
		})
	}
}
