/*
Copyright 2021 The kconnect Authors.

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

package identity

import (
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
)

// ExplicitBearerAuthorizer implements a bearer token authorizer where the token is supplied at
// creation and not via a token provider
type ExplicitBearerAuthorizer struct {
	token string
}

// NewExplicitBearerAuthorizer creates a new ExplicitBearerAuthorizer with the specified token.
func NewExplicitBearerAuthorizer(token string) *ExplicitBearerAuthorizer {
	return &ExplicitBearerAuthorizer{
		token: token,
	}
}

// WithAuthorization returns a PrepareDecorator that adds an HTTP Authorization header whose
// value is "Bearer " followed by the bearer token
func (ba *ExplicitBearerAuthorizer) WithAuthorization() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err == nil {
				return autorest.Prepare(r, autorest.WithHeader("Authorization", fmt.Sprintf("Bearer %s", ba.token)))
			}
			return r, err
		})
	}
}
