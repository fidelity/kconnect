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

package defaults

import (
	"fmt"

	"github.com/fidelity/kconnect/internal/version"
	"github.com/fidelity/kconnect/pkg/http"
)

type Option func(map[string]string)

func Headers(opts ...Option) map[string]string {
	v := version.Get()

	headers := make(map[string]string)
	headers["User-Agent"] = fmt.Sprintf("kconnect/%s", v.Version)

	for _, opt := range opts {
		opt(headers)
	}

	return headers
}

func WithNoCache() Option {
	return func(headers map[string]string) {
		headers["Cache-Control"] = "no-cache"
	}
}

func WithJSON() Option {
	return func(headers map[string]string) {
		headers["Accept"] = http.MediaTypeJSON
		headers["Content-Type"] = http.MediaTypeJSON
	}
}

func WithContentTypeJSON() Option {
	return func(headers map[string]string) {
		headers["Content-Type"] = http.MediaTypeJSON
	}
}

func WithAcceptJSON() Option {
	return func(headers map[string]string) {
		headers["Accept"] = http.MediaTypeJSON
	}
}

func WithBearerAuth(token string) Option {
	return func(headers map[string]string) {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}
}
