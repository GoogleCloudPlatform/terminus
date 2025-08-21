// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terminus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPRequestMsg is sent when an HTTP request completes
type HTTPRequestMsg struct {
	Response *http.Response
	Body     []byte
	Error    error
	Tag      string // Optional tag to identify the request
}

// HTTPMethod represents an HTTP method
type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
)

// HTTPRequest performs an HTTP request and returns the result as a message
func HTTPRequest(method HTTPMethod, url string, body io.Reader) Cmd {
	return HTTPRequestWithContext(context.Background(), method, url, body, nil, "")
}

// HTTPRequestWithTag performs an HTTP request with a tag for identification
func HTTPRequestWithTag(method HTTPMethod, url string, body io.Reader, tag string) Cmd {
	return HTTPRequestWithContext(context.Background(), method, url, body, nil, tag)
}

// HTTPRequestWithHeaders performs an HTTP request with custom headers
func HTTPRequestWithHeaders(method HTTPMethod, url string, body io.Reader, headers map[string]string) Cmd {
	return HTTPRequestWithContext(context.Background(), method, url, body, headers, "")
}

// HTTPRequestWithContext performs an HTTP request with a context for cancellation
func HTTPRequestWithContext(ctx context.Context, method HTTPMethod, url string, body io.Reader, headers map[string]string, tag string) Cmd {
	return func() Msg {
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequestWithContext(ctx, string(method), url, body)
		if err != nil {
			return HTTPRequestMsg{
				Error: fmt.Errorf("failed to create request: %w", err),
				Tag:   tag,
			}
		}

		// Set default headers
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// Set custom headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			return HTTPRequestMsg{
				Error: fmt.Errorf("request failed: %w", err),
				Tag:   tag,
			}
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return HTTPRequestMsg{
				Response: resp,
				Error:    fmt.Errorf("failed to read response body: %w", err),
				Tag:      tag,
			}
		}

		return HTTPRequestMsg{
			Response: resp,
			Body:     bodyBytes,
			Error:    nil,
			Tag:      tag,
		}
	}
}

// JSONRequest performs an HTTP request with JSON encoding/decoding
func JSONRequest(method HTTPMethod, url string, data interface{}) Cmd {
	return JSONRequestWithTag(method, url, data, "")
}

// JSONRequestWithTag performs an HTTP request with JSON encoding/decoding and a tag
func JSONRequestWithTag(method HTTPMethod, url string, data interface{}, tag string) Cmd {
	var body io.Reader
	if data != nil {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return func() Msg {
				return HTTPRequestMsg{
					Error: fmt.Errorf("failed to marshal JSON: %w", err),
					Tag:   tag,
				}
			}
		}
		body = bytes.NewReader(jsonBytes)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	return HTTPRequestWithHeaders(method, url, body, headers)
}

// Get performs a GET request
func Get(url string) Cmd {
	return HTTPRequest(GET, url, nil)
}

// GetWithTag performs a GET request with a tag
func GetWithTag(url string, tag string) Cmd {
	return HTTPRequestWithTag(GET, url, nil, tag)
}

// Post performs a POST request with JSON data
func Post(url string, data interface{}) Cmd {
	return JSONRequest(POST, url, data)
}

// PostWithTag performs a POST request with JSON data and a tag
func PostWithTag(url string, data interface{}, tag string) Cmd {
	return JSONRequestWithTag(POST, url, data, tag)
}

// Put performs a PUT request with JSON data
func Put(url string, data interface{}) Cmd {
	return JSONRequest(PUT, url, data)
}

// Delete performs a DELETE request
func Delete(url string) Cmd {
	return HTTPRequest(DELETE, url, nil)
}

// IsHTTPError checks if the HTTP response indicates an error
func (msg HTTPRequestMsg) IsHTTPError() bool {
	return msg.Response != nil && msg.Response.StatusCode >= 400
}

// IsNetworkError checks if the request failed due to a network error
func (msg HTTPRequestMsg) IsNetworkError() bool {
	return msg.Error != nil && msg.Response == nil
}

// StatusCode returns the HTTP status code, or 0 if no response
func (msg HTTPRequestMsg) StatusCode() int {
	if msg.Response != nil {
		return msg.Response.StatusCode
	}
	return 0
}

// JSONBody attempts to unmarshal the response body as JSON
func (msg HTTPRequestMsg) JSONBody(v interface{}) error {
	if msg.Error != nil {
		return msg.Error
	}
	if len(msg.Body) == 0 {
		return fmt.Errorf("empty response body")
	}
	return json.Unmarshal(msg.Body, v)
}

// String returns the response body as a string
func (msg HTTPRequestMsg) String() string {
	return string(msg.Body)
}