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

package activedirectory

import "encoding/json"

type loginRequest struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type loginResponse struct { //TODO: add additional fields
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Token  string      `json:"token"`
	UserID string      `json:"userId"`
	TTL    json.Number `json:"ttl"`
}
