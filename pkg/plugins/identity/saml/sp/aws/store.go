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
	"fmt"

	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/spf13/pflag"
	"github.com/versent/saml2aws/pkg/awsconfig"
)

// NewIdentityStore will create a new AWS identity store
func NewIdentityStore(providerFlags *pflag.FlagSet) (provider.IdentityStore, error) {
	if !flags.ExistsWithValue("profile", providerFlags) {
		return nil, ErrNoProfile
	}
	profileFlag := providerFlags.Lookup("profile")

	return &awsIdentityStore{
		configProvider: awsconfig.NewSharedCredentials(profileFlag.Value.String()),
	}, nil
}

type awsIdentityStore struct {
	configProvider *awsconfig.CredentialsProvider
}

func (s *awsIdentityStore) CredsExists() (bool, error) {
	return s.configProvider.CredsExists()
}

func (s *awsIdentityStore) Save(identity provider.Identity) error {
	awsIdentity, ok := identity.(*Identity)
	if !ok {
		return fmt.Errorf("expected AWSIdentity but got a %T: %w", identity, ErrUnexpectedIdentity)
	}
	awsCreds := mapIdentityToCreds(awsIdentity)

	return s.configProvider.Save(awsCreds)
}

func (s *awsIdentityStore) Load() (provider.Identity, error) {
	creds, err := s.configProvider.Load()
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}
	awsID := mapCredsToIdentity(creds, s.configProvider.Profile)

	return awsID, nil
}

func (s *awsIdentityStore) Expired() bool {
	return s.configProvider.Expired()
}
