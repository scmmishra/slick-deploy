package caddy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a test server that returns a 200 status code
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a new CaddyClient
	client := NewCaddyClient(server.URL)

	// Test the Load function
	err := client.Load("test caddyfile")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestLoad_ErrorConditions(t *testing.T) {
	// Test when the request creation fails
	{
		client := &CaddyClient{
			BaseURL:    ":", // invalid URL
			HTTPClient: &http.Client{},
		}
		err := client.Load("test caddyfile")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}

	// Test when the HTTP client fails to send the request
	{
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
		}))
		server.Close() // close the server to simulate network error
		client := NewCaddyClient(server.URL)
		err := client.Load("test caddyfile")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}

	// Test when the server returns a non-200 status code
	{
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusInternalServerError) // return 500 status code
		}))
		defer server.Close()
		client := NewCaddyClient(server.URL)
		err := client.Load("test caddyfile")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}
}
