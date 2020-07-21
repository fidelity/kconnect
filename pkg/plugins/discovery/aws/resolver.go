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
package aws

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	idprov "github.com/fidelity/kconnect/pkg/plugins/identity/saml"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/spf13/pflag"
)

var (
	errNoRoleArnFlag = errors.New("no role-arn flag found in resolver")
)

// FlagsResolver is used to resolve the values for AWS related flags interactively.
type FlagsResolver struct {
	id      *idprov.AWSIdentity
	session awsclient.ConfigProvider
}

// Resolve will resolve the values for the AWS specific flags that have no value. It will
// query AWS and interactively ask the user for selections.
func (r *FlagsResolver) Resolve(identity provider.Identity, flags *pflag.FlagSet) error {
	r.id = identity.(*idprov.AWSIdentity)

	session, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-2"),
		Credentials: credentials.NewSharedCredentials("", r.id.Profile()),
	})
	if err != nil {
		return fmt.Errorf("getting aws session: %w", err)
	}
	r.session = session

	if err := r.resolveRoleArn(flags); err != nil {
		return fmt.Errorf("resolving role-arn: %w", err)
	}

	return nil
}

func (r *FlagsResolver) resolveRoleArn(flags *pflag.FlagSet) error {
	flag := flags.Lookup("role-arn")
	if flag == nil {
		return errNoRoleArnFlag
	}
	if flag.Value.String() != "" {
		return nil
	}

	iamClient := newIAMClient(r.session)

	input := &iam.ListRolesInput{}

	roles := []*iam.Role{}
	err := iamClient.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, role := range page.Roles {
			//TODO: apply the filter
			roles = append(roles, role)
		}
		return true
	})
	if err != nil {
		return fmt.Errorf("listing iam roles: %w", err)
	}

	//TODO: prompt user

	return nil
}
