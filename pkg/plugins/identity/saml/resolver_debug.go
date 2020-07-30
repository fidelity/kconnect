// +build debug

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

	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/spf13/pflag"
)

func (r *awsFlagsResolver) resolveProfile(name string, flagset *pflag.FlagSet) error {
	if flags.ExistsWithValue(name, flagset) {
		return nil
	}

	profile := "samltest"

	if err := flagset.Set(name, profile); err != nil {
		r.logger.Errorf("failed setting profile flag to %s: %s", profile, err.Error())
		return fmt.Errorf("setting profile flag: %w", err)
	}

	return nil
}

func (r *awsFlagsResolver) resolveRegion(name string, flagset *pflag.FlagSet) error {
	if flags.ExistsWithValue(name, flagset) {
		return nil
	}

	region := "us-west-2"

	if err := flagset.Set(name, region); err != nil {
		r.logger.Errorf("failed setting region flag to %s: %s", region, err.Error())
		return fmt.Errorf("setting region flag: %w", err)
	}

	return nil
}

func (r *awsFlagsResolver) resolveUsername(name string, flagset *pflag.FlagSet) error {
	if flags.ExistsWithValue(name, flagset) {
		return nil
	}

	username := "testuser"

	if err := flagset.Set(name, username); err != nil {
		r.logger.Errorf("failed setting username flag to %s: %s", username, err.Error())
		return fmt.Errorf("setting username flag: %w", err)
	}

	return nil
}

func (r *awsFlagsResolver) resolvePassword(name string, flagset *pflag.FlagSet) error {
	if flags.ExistsWithValue(name, flagset) {
		return nil
	}

	//TODO: change to blank out the characters
	password := "password"

	if err := flagset.Set(name, password); err != nil {
		r.logger.Errorf("failed setting password flag to %s: %s", password, err.Error())
		return fmt.Errorf("setting password flag: %w", err)
	}

	return nil
}

func (r *awsFlagsResolver) resolveIdpEndpoint(name string, flagset *pflag.FlagSet) error {
	if flags.ExistsWithValue(name, flagset) {
		return nil
	}

	endpoint := "http://someendpoint"

	if err := flagset.Set(name, endpoint); err != nil {
		r.logger.Errorf("failed setting idp-endpoint flag to %s: %s", endpoint, err.Error())
		return fmt.Errorf("setting idp-endpoint flag: %w", err)
	}

	return nil
}

func (r *awsFlagsResolver) resolveIdpProvider(name string, flagset *pflag.FlagSet) error {
	if flags.ExistsWithValue(name, flagset) {
		return nil
	}

	idpProvider := "GoogleApps"

	if err := flagset.Set(name, idpProvider); err != nil {
		r.logger.Errorf("failed setting idp-provider flag to %s: %s", idpProvider, err.Error())
		return fmt.Errorf("setting idp-provider flag: %w", err)
	}

	return nil
}

// func (r *awsFlagsResolver) resolveRoleARN(name string, flagset *pflag.FlagSet) error {
// 	if flags.ExistsWithValue(name, flagset) {
// 		return nil
// 	}

// 	roleArn := "some role"

// 	if err := flagset.Set(name, roleArn); err != nil {
// 		r.logger.Errorf("failed setting role-arn flag to %s: %s", role-arn, err.Error())
// 		return fmt.Errorf("setting role-arn flag: %w", err)
// 	}

// 	return nil
// }
