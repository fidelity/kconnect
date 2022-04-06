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
	"time"
)

// Identity represents an AWS identity
type Identity struct {
	ProfileName              string
	AWSAccessKey             string
	AWSSecretKey             string
	AWSSessionToken          string
	AWSSecurityToken         string
	PrincipalARN             string
	Expires                  time.Time
	Region                   string
	AWSSharedCredentialsFile string
	IDProviderName           string
}

func (i *Identity) Type() string {
	return "aws"
}

func (i *Identity) Name() string {
	return i.PrincipalARN
}

func (i *Identity) IsExpired() bool {
	now := time.Now().UTC()
	return now.After(i.Expires)
}

func (i *Identity) IdentityProviderName() string {
	return i.IDProviderName
}
