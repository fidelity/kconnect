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

package use

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/fidelity/kconnect/pkg/app"
)

func TestGetIdpProtocolFromFlagAndEnvironment(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		env           string
		expectValue   string
		expectHasFlag bool
	}{
		{
			name:          "flag with a space",
			args:          []string{"kconnect", "use", "eks", "--idp-protocol", "saml"},
			expectValue:   "saml",
			expectHasFlag: true,
		},
		{
			name:          "flag with an equals",
			args:          []string{"kconnect", "use", "eks", "--idp-protocol=saml"},
			expectValue:   "saml",
			expectHasFlag: true,
		},
		{
			name:          "environment variable",
			args:          []string{"kconnect", "use", "eks"},
			env:           "entraid-oauth",
			expectValue:   "entraid-oauth",
			expectHasFlag: false,
		},
		{
			name:          "flag takes priority over the environment variable",
			args:          []string{"kconnect", "use", "eks", "--idp-protocol=saml"},
			env:           "aws-iam",
			expectValue:   "saml",
			expectHasFlag: true,
		},
		{
			name:          "flag with no value falls back to the environment variable",
			args:          []string{"kconnect", "use", "eks", "--idp-protocol"},
			env:           "aws-iam",
			expectValue:   "aws-iam",
			expectHasFlag: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			t.Setenv(idpProtocolEnvVar, tc.env)

			value, hasFlag, err := getIdpProtocol(tc.args, &app.UseInput{DiscoveryProvider: "eks"})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(value).To(Equal(tc.expectValue))
			g.Expect(hasFlag).To(Equal(tc.expectHasFlag))
		})
	}
}
