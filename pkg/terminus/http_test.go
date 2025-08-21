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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "hello"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name       string
		method     HTTPMethod
		path       string
		wantStatus int
		wantBody   string
		wantError  bool
	}{
		{
			name:       "GET success",
			method:     GET,
			path:       "/success",
			wantStatus: http.StatusOK,
			wantBody:   "success",
		},
		{
			name:       "GET error",
			method:     GET,
			path:       "/error",
			wantStatus: http.StatusInternalServerError,
			wantBody:   "error",
		},
		{
			name:       "POST success",
			method:     POST,
			path:       "/success",
			wantStatus: http.StatusOK,
			wantBody:   "success",
		},
		{
			name:       "Invalid URL",
			method:     GET,
			path:       "://invalid",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + tt.path
			if tt.wantError {
				url = tt.path // Use invalid URL directly
			}

			cmd := HTTPRequest(tt.method, url, nil)
			msg := cmd()

			httpMsg, ok := msg.(HTTPRequestMsg)
			if !ok {
				t.Fatal("Expected HTTPRequestMsg")
			}

			if tt.wantError {
				if httpMsg.Error == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if httpMsg.Error != nil {
				t.Errorf("Unexpected error: %v", httpMsg.Error)
			}

			if httpMsg.StatusCode() != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, httpMsg.StatusCode())
			}

			if string(httpMsg.Body) != tt.wantBody {
				t.Errorf("Expected body %q, got %q", tt.wantBody, string(httpMsg.Body))
			}
		})
	}
}

func TestHTTPRequestWithTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tag := "test-request"
	cmd := HTTPRequestWithTag(GET, server.URL, nil, tag)
	msg := cmd()

	httpMsg, ok := msg.(HTTPRequestMsg)
	if !ok {
		t.Fatal("Expected HTTPRequestMsg")
	}

	if httpMsg.Tag != tag {
		t.Errorf("Expected tag %q, got %q", tag, httpMsg.Tag)
	}
}

func TestJSONRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Echo back the JSON
		var data map[string]interface{}
		json.NewDecoder(r.Body).Decode(&data)
		json.NewEncoder(w).Encode(data)
	}))
	defer server.Close()

	testData := map[string]string{"key": "value"}
	cmd := JSONRequest(POST, server.URL, testData)
	msg := cmd()

	httpMsg, ok := msg.(HTTPRequestMsg)
	if !ok {
		t.Fatal("Expected HTTPRequestMsg")
	}

	if httpMsg.Error != nil {
		t.Fatalf("Unexpected error: %v", httpMsg.Error)
	}

	var response map[string]string
	err := httpMsg.JSONBody(&response)
	if err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if response["key"] != "value" {
		t.Errorf("Expected key=value, got %v", response)
	}
}

func TestHTTPHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the method back
		w.Write([]byte(r.Method))
	}))
	defer server.Close()

	tests := []struct {
		name   string
		cmd    Cmd
		method string
	}{
		{"GET", Get(server.URL), "GET"},
		{"POST", Post(server.URL, nil), "POST"},
		{"PUT", Put(server.URL, nil), "PUT"},
		{"DELETE", Delete(server.URL), "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.cmd()
			httpMsg, ok := msg.(HTTPRequestMsg)
			if !ok {
				t.Fatal("Expected HTTPRequestMsg")
			}

			if string(httpMsg.Body) != tt.method {
				t.Errorf("Expected method %s, got %s", tt.method, string(httpMsg.Body))
			}
		})
	}
}

func TestHTTPRequestMsgHelpers(t *testing.T) {
	t.Run("IsHTTPError", func(t *testing.T) {
		msg := HTTPRequestMsg{
			Response: &http.Response{StatusCode: http.StatusNotFound},
		}
		if !msg.IsHTTPError() {
			t.Error("Expected IsHTTPError to return true for 404")
		}

		msg.Response.StatusCode = http.StatusOK
		if msg.IsHTTPError() {
			t.Error("Expected IsHTTPError to return false for 200")
		}
	})

	t.Run("IsNetworkError", func(t *testing.T) {
		msg := HTTPRequestMsg{
			Error: http.ErrHandlerTimeout,
		}
		if !msg.IsNetworkError() {
			t.Error("Expected IsNetworkError to return true when error and no response")
		}

		msg.Response = &http.Response{}
		if msg.IsNetworkError() {
			t.Error("Expected IsNetworkError to return false when response exists")
		}
	})
}

func TestHTTPTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Use Timeout command
	cmd := Timeout(50*time.Millisecond, Get(server.URL))
	msg := cmd()

	if _, ok := msg.(TimeoutMsg); !ok {
		t.Error("Expected TimeoutMsg for slow request")
	}
}