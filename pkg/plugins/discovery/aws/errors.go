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

import "errors"

var (
	ErrNoRoleArnFlag           = errors.New("no role-arn flag found in resolver")
	ErrNoSession               = errors.New("no aws session supplied")
	ErrFlagMissing             = errors.New("flag missing")
	ErrNotAWSIdentity          = errors.New("unsupported identity, AWSIdentity required")
	ErrUnexpectedClusterFormat = errors.New("cluster name from ARN has unexpected format")
)
