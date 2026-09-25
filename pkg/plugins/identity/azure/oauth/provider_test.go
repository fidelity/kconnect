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

package oauth

import (
	"reflect"
	"testing"
)

func Test_rolesFromAWSRolesClaim(t *testing.T) {
	testCases := []struct {
		name      string
		rawRoles  any
		wantRoles []awsRole
		wantErr   bool
	}{
		{
			name:     "single role as bare string (Entra ID single-role shape)",
			rawRoles: "arn:aws:iam::123456789012:role/my-role",
			wantRoles: []awsRole{
				{Name: "123456789012 / my-role", RoleARN: "arn:aws:iam::123456789012:role/my-role"},
			},
		},
		{
			name:     "single role as bare string with SAML provider suffix",
			rawRoles: "arn:aws:iam::123456789012:role/my-role,arn:aws:iam::123456789012:saml-provider/my-provider",
			wantRoles: []awsRole{
				{Name: "123456789012 / my-role", RoleARN: "arn:aws:iam::123456789012:role/my-role"},
			},
		},
		{
			name: "multiple roles as array",
			rawRoles: []any{
				"arn:aws:iam::123456789012:role/role-a",
				"arn:aws:iam::123456789012:role/role-b",
			},
			wantRoles: []awsRole{
				{Name: "123456789012 / role-a", RoleARN: "arn:aws:iam::123456789012:role/role-a"},
				{Name: "123456789012 / role-b", RoleARN: "arn:aws:iam::123456789012:role/role-b"},
			},
		},
		{
			name:     "unexpected type",
			rawRoles: 123,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roles, err := rolesFromAWSRolesClaim(tc.rawRoles)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(roles, tc.wantRoles) {
				t.Errorf("got roles %+v, want %+v", roles, tc.wantRoles)
			}
		})
	}
}
