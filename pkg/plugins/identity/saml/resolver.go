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
package saml

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/fidelity/kconnect/pkg/aws"
	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/provider"
)

// NewAWSFlagsResolver creates a new flags resolver for AWS
func NewAWSFlagsResolver(logger *logrus.Entry) provider.FlagsResolver {
	return &awsFlagsResolver{
		logger: logger,
	}
}

// awsFlagsResolver is used to resolve the values for AWS related flags interactively.
type awsFlagsResolver struct {
	logger  *logrus.Entry
	session client.ConfigProvider
}

func (r *awsFlagsResolver) Validate(ctx *provider.Context, flagset *pflag.FlagSet) error {
	errsValidation := &kerrors.ValidationFailed{}

	//TODO: create a declarative way to define flags

	// Region
	if flags.ExistsWithValue("region", flagset) == false {
		errsValidation.AddFailure("region is required")
	}

	// Profile
	if flags.ExistsWithValue("profile", flagset) == false {
		errsValidation.AddFailure("profile is required")
	}

	// Username
	if flags.ExistsWithValue("username", flagset) == false {
		errsValidation.AddFailure("username is required")
	}

	// Password
	if flags.ExistsWithValue("password", flagset) == false {
		errsValidation.AddFailure("password is required")
	}

	// IDPEndpoint
	if flags.ExistsWithValue("idp-endpoint", flagset) == false {
		errsValidation.AddFailure("idp-endpoint is required")
	}

	if len(errsValidation.Failures()) > 0 {
		return errsValidation
	}

	return nil
}

// Resolve will resolve the values for the AWS specific flags that have no value. It will
// query AWS and interactively ask the user for selections.
func (r *awsFlagsResolver) Resolve(ctx *provider.Context, flagset *pflag.FlagSet) error {
	logger := ctx.Logger().WithField("resolver", "aws-identity")
	logger.Debug("resolving AWS identity flags")

	// TODO: handle in declarative mannor
	// NOTE: resolution is only needed for required fields
	if err := r.resolveIdpProvider("idp-provider", flagset); err != nil {
		return fmt.Errorf("resolving idp-provider: %w", err)
	}
	if err := r.resolveIdpEndpoint("idp-endpoint", flagset); err != nil {
		return fmt.Errorf("resolving idp-endpoint: %w", err)
	}
	if err := r.resolveProfile("profile", flagset); err != nil {
		return fmt.Errorf("resolving profile: %w", err)
	}
	if err := r.resolveRegion("region", flagset); err != nil {
		return fmt.Errorf("resolving region: %w", err)
	}
	if err := r.resolveUsername("username", flagset); err != nil {
		return fmt.Errorf("resolving username: %w", err)
	}
	if err := r.resolvePassword("password", flagset); err != nil {
		return fmt.Errorf("resolving password: %w", err)
	}

	// if err := r.resolveRoleARN("role-arn", flagset); err != nil {
	// 	return fmt.Errorf("resolving role-arn: %w", err)
	// }

	return nil
}

func (r *awsFlagsResolver) ensureSessionFromFlags(flagset *pflag.FlagSet) error {
	if r.session != nil {
		return nil
	}

	if !flags.ExistsWithValue("region", flagset) {
		return ErrNoRegion

	}
	if !flags.ExistsWithValue("profile", flagset) {
		return ErrNoProfile
	}

	region, err := flagset.GetString("region")
	if err != nil {
		return fmt.Errorf("getting region flag: %w", err)
	}
	profile, err := flagset.GetString("profile")
	if err != nil {
		return fmt.Errorf("getting profile flag: %w", err)
	}

	session, err := aws.NewSession(region, profile)
	if err != nil {
		return fmt.Errorf("creating new AWS session: %w", err)
	}

	r.session = session

	return nil
}
