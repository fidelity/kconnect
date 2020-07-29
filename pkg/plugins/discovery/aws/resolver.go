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

	awsclient "github.com/aws/aws-sdk-go/aws/client"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	kerrors "github.com/fidelity/kconnect/pkg/errors"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/provider"
)

var (
	ErrNoRoleArnFlag = errors.New("no role-arn flag found in resolver")
	ErrNoSession     = errors.New("no aws session supplied")
	ErrFlagMissing   = errors.New("flag missing")
)

// NewFlagsResolver creates a new flags resolver for AWS
func NewFlagsResolver(session awsclient.ConfigProvider, logger *logrus.Entry) (provider.FlagsResolver, error) {
	if session == nil {
		return nil, ErrNoSession
	}

	return &awsFlagsResolver{
		session: session,
		logger:  logger,
	}, nil
}

// awsFlagsResolver is used to resolve the values for AWS related flags interactively.
type awsFlagsResolver struct {
	session awsclient.ConfigProvider
	logger  *logrus.Entry
}

func (r *awsFlagsResolver) Validate(ctx *provider.Context, flagset *pflag.FlagSet) error {
	errsValidation := &kerrors.ValidationFailed{}

	//TODO: create a declarative way to define flags

	// Region
	if !flags.ExistsWithValue("role-arn", flagset) {
		errsValidation.AddFailure("role-arn is required")
	}

	if len(errsValidation.Failures()) > 0 {
		return errsValidation
	}

	return nil
}

// Resolve will resolve the values for the AWS specific flags that have no value. It will
// query AWS and interactively ask the user for selections.
func (r *awsFlagsResolver) Resolve(ctx *provider.Context, flagset *pflag.FlagSet) error {
	logger := ctx.Logger().WithField("resolver", "aws")
	logger.Debug("resolving AWS flags")

	//TODO: handle in declarative mannor

	return nil
}
