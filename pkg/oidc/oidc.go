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

package oidc

import "time"

// Identity is an identity that is based on OpenID Connect
type Identity struct {
	Scope        string `json:"scope"`
	Expires      time.Time
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

func (i *Identity) Type() string {
	return "oidc"
}

func (i *Identity) Name() string {
	return "" //TODO: get from ID token?
}

func (i *Identity) IsExpired() bool {
	now := time.Now().UTC()
	return now.After(i.Expires)
}
