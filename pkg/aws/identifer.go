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
	"hash/fnv"

	"github.com/versent/saml2aws/v2/pkg/awsconfig"
)

var (
	ErrPrincipleARNRequired = errors.New("principle arn is required ")
)

// CreateIDFromCreds will create a unique identifier for a set of credentials. The identifier
// is a hash of the principle ARN
func CreateIDFromCreds(creds *awsconfig.AWSCredentials) (string, error) {
	if creds.PrincipalARN == "" {
		return "", ErrPrincipleARNRequired
	}
	h := fnv.New32a()
	if _, err := h.Write([]byte(creds.PrincipalARN)); err != nil {
		return "", fmt.Errorf("write to has function: %w", err)
	}

	return fmt.Sprintf("%d", h.Sum32()), nil
}
