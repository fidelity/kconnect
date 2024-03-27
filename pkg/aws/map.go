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

import "github.com/versent/saml2aws/v2/pkg/awsconfig"

func MapCredsToIdentity(creds *awsconfig.AWSCredentials, profileName, awsSharedCredentialsFile string) *Identity {
	return &Identity{
		AWSAccessKey:             creds.AWSAccessKey,
		AWSSecretKey:             creds.AWSSecretKey,
		AWSSecurityToken:         creds.AWSSecurityToken,
		AWSSessionToken:          creds.AWSSessionToken,
		AWSSharedCredentialsFile: awsSharedCredentialsFile,
		Expires:                  creds.Expires,
		PrincipalARN:             creds.PrincipalARN,
		ProfileName:              profileName,
		Region:                   creds.Region,
	}
}

func MapIdentityToCreds(awsIdentity *Identity) *awsconfig.AWSCredentials {
	return &awsconfig.AWSCredentials{
		AWSAccessKey:     awsIdentity.AWSAccessKey,
		AWSSecretKey:     awsIdentity.AWSSecretKey,
		AWSSecurityToken: awsIdentity.AWSSecurityToken,
		AWSSessionToken:  awsIdentity.AWSSessionToken,
		Expires:          awsIdentity.Expires,
		PrincipalARN:     awsIdentity.PrincipalARN,
		Region:           awsIdentity.Region,
	}
}
