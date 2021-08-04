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

	"github.com/versent/saml2aws/v2/pkg/awsconfig"

	"github.com/fidelity/kconnect/pkg/provider/identity"

	kawsconfig "github.com/fidelity/kconnect/pkg/aws/awsconfig"
)

// NewIdentityStore will create a new AWS identity store
func NewIdentityStore(profile, idProviderName string) (identity.Store, error) {
	path, nil := kawsconfig.LocateConfigFile()
	return &awsIdentityStore{
		configProvider: awsconfig.NewSharedCredentials(profile, path),
		idProviderName: idProviderName,
	}, nil
}

type awsIdentityStore struct {
	configProvider *awsconfig.CredentialsProvider
	idProviderName string
}

func (s *awsIdentityStore) CredsExists() (bool, error) {
	return s.configProvider.CredsExists()
}

func (s *awsIdentityStore) Save(userID identity.Identity) error {
	awsIdentity, ok := userID.(*Identity)
	if !ok {
		return fmt.Errorf("expected AWSIdentity but got a %T: %w", userID, ErrUnexpectedIdentity)
	}
	awsCreds := MapIdentityToCreds(awsIdentity)

	return s.configProvider.Save(awsCreds)
}

func (s *awsIdentityStore) Load() (identity.Identity, error) {
	creds, err := s.configProvider.Load()
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}
	awsID := MapCredsToIdentity(creds, s.configProvider.Profile)

	return awsID, nil
}

func (s *awsIdentityStore) Expired() bool {
	return s.configProvider.Expired()
}
