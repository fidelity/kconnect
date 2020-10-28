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

package http

// Client represents an http client
type Client interface {
	Do(req *ClientRequest) (ClientResponse, error)
	Get(url string, headers map[string]string) (ClientResponse, error)
	Post(url string, body string, headers map[string]string) (ClientResponse, error)
}

// ClientRequest represents a http request
type ClientRequest struct {
	URL     string
	Body    string
	Method  string
	Headers map[string]string
}

// ClientResponse represents a http client response
type ClientResponse interface {
	ResponseCode() int
	Body() string
	Headers() map[string]string
}
