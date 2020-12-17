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

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

const (
	MediaTypeJSON = "application/json"
)

// NewHTTPClient creates a new http client
func NewHTTPClient() Client {
	client := &http.Client{}

	return &netHTTPClient{client}
}

// netHttpClient is a http client based on net/http
type netHTTPClient struct {
	client *http.Client
}

func (n *netHTTPClient) Do(req *ClientRequest) (ClientResponse, error) {
	var r *http.Request
	var err error

	if req.Body == nil {
		r, err = http.NewRequest(req.Method, req.URL, nil)
	} else {
		r, err = http.NewRequest(req.Method, req.URL, strings.NewReader(*req.Body))
	}
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}
	for k, v := range req.Headers {
		r.Header.Add(k, v)
	}

	zap.S().Debugw("http request", "url", req.URL, "method", req.Method, "headers", req.Headers)

	resp, err := n.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return createResponse(resp)

}

func (n *netHTTPClient) Get(url string, headers map[string]string) (ClientResponse, error) {
	req := &ClientRequest{
		Method:  "GET",
		Headers: headers,
		URL:     url,
	}

	return n.Do(req)
}

func (n *netHTTPClient) Post(url string, body string, headers map[string]string) (ClientResponse, error) {
	req := &ClientRequest{
		Method:  "POST",
		Headers: headers,
		Body:    &body,
		URL:     url,
	}

	return n.Do(req)
}

func createResponse(resp *http.Response) (ClientResponse, error) {
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("getting response body: %w", err)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = v[0]
	}

	zap.S().Debugw("http response", "status", resp.Status, "headers", headers)
	zap.S().Debug(string(body))

	return &netHTTPResponse{
		code:    resp.StatusCode,
		body:    string(body),
		headers: headers,
	}, nil
}

type netHTTPResponse struct {
	code    int
	body    string
	headers map[string]string
}

func (r *netHTTPResponse) ResponseCode() int {
	return r.code
}

func (r *netHTTPResponse) Body() string {
	return r.body
}

func (r *netHTTPResponse) Headers() map[string]string {
	return r.headers
}
