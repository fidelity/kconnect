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

package http_test

import (
	"fmt"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	khttp "github.com/fidelity/kconnect/pkg/http"
)

func TestRedactHeaders(t *testing.T) {
	g := NewWithT(t)

	headers := map[string]string{
		"Authorization": "Bearer a-token",
		"cookie":        "session=abc",
		"Set-Cookie":    "session=abc",
		"X-API-Key":     "a-key",
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}

	redacted := khttp.RedactHeaders(headers)

	g.Expect(redacted).To(Equal(map[string]string{
		"Authorization": "********",
		"cookie":        "********",
		"Set-Cookie":    "********",
		"X-API-Key":     "********",
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}))
	// The headers that are sent must not be changed
	g.Expect(headers["Authorization"]).To(Equal("Bearer a-token"))
	g.Expect(khttp.RedactHeaders(nil)).To(BeEmpty())
}

func TestCredentialsAreNotLogged(t *testing.T) {
	g := NewWithT(t)

	core, logs := observer.New(zapcore.DebugLevel)
	restore := zap.ReplaceGlobals(zap.New(core))
	t.Cleanup(restore)

	var receivedAuth string

	server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		receivedAuth = r.Header.Get("Authorization")

		w.Header().Set("Set-Cookie", "session=response-cookie-secret")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"response-token-secret"}`)
	}))
	t.Cleanup(server.Close)

	headers := map[string]string{"Accept": "application/json"}
	khttp.SetBasicAuthHeaders(headers, "pete", "request-password-secret")

	resp, err := khttp.NewHTTPClient().Post(server.URL, `{"password":"body-password-secret"}`, headers)
	g.Expect(err).NotTo(HaveOccurred())

	// The request and the response are unchanged
	g.Expect(receivedAuth).To(HavePrefix("Basic "))
	g.Expect(resp.ResponseCode()).To(Equal(200))
	g.Expect(resp.Body()).To(ContainSubstring("response-token-secret"))

	logged := []string{}
	for _, entry := range logs.All() {
		logged = append(logged, entry.Message+fmt.Sprintf(" %v", entry.ContextMap()))
	}

	all := strings.Join(logged, "\n")
	g.Expect(logs.Len()).To(BeNumerically(">=", 2))
	g.Expect(all).To(ContainSubstring("application/json"))
	g.Expect(all).NotTo(ContainSubstring(receivedAuth))
	g.Expect(all).NotTo(ContainSubstring("response-cookie-secret"))
	g.Expect(all).NotTo(ContainSubstring("response-token-secret"))
	g.Expect(all).NotTo(ContainSubstring("body-password-secret"))
}
