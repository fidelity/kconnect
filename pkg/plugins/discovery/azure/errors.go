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

package azure

import "errors"

var (
	ErrUnsupportedIdentity  = errors.New("unsupported identity, oidc.Identity orazure.AuthorizerIdentity required")
	ErrNoKubeconfigs        = errors.New("no kubeconfigs available for the managed cluster cluster")
	ErrNoSubscriptions      = errors.New("no subscriptions found")
	ErrSubscriptionNameOrID = errors.New("subscription name and id cannot be both supplied")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrTokenNeedsAD         = errors.New("the 'token' login type requires using aad idp-protocol")
)
