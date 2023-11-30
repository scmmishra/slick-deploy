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
