package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/stretchr/testify/assert"
)

// setupTestServer helps in creating a test HTTP server.
func setupTestServer(handler http.HandlerFunc) (string, func()) {
	ts := httptest.NewServer(handler)
	return ts.URL, func() { ts.Close() }
}

func TestCheckHealth_EmptyHost(t *testing.T) {
	t.Parallel()

	err := CheckHealth("", &config.HealthCheck{
		Endpoint:       "/health",
		TimeoutSeconds: 5,
	})

	assert.NoError(t, err, "Expected no error for healthy response")
}

func TestCheckHealth_EmptyEndpoint(t *testing.T) {
	t.Parallel()

	err := CheckHealth("https://shivam.dev", &config.HealthCheck{
		Endpoint:       "",
		TimeoutSeconds: 5,
	})

	assert.NoError(t, err, "Expected no error for healthy response")
}

// TestCheckHealth_Success tests successful health check.
func TestCheckHealth_Success(t *testing.T) {
	t.Parallel()

	serverURL, teardown := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Return 200 OK
	})
	defer teardown()

	err := CheckHealth(serverURL, &config.HealthCheck{
		Endpoint:       "/health",
		TimeoutSeconds: 5,
	})

	assert.NoError(t, err, "Expected no error for healthy response")
}

// TestCheckHealth_ServerError tests the health check for server error response.
func TestCheckHealth_ServerError(t *testing.T) {
	t.Parallel()

	serverURL, teardown := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // Return 500 Internal Server Error
	})
	defer teardown()

	err := CheckHealth(serverURL, &config.HealthCheck{
		Endpoint:       "/health",
		TimeoutSeconds: 5,
	})

	assert.Error(t, err, "Expected an error for server error response")
}

// TestCheckHealth_NetworkError tests the health check for a network error.
func TestCheckHealth_NetworkError(t *testing.T) {
	t.Parallel()

	serverURL, teardown := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		// You can configure the response here if needed
	})
	// Close the server immediately to simulate a network error
	teardown()

	err := CheckHealth(serverURL, &config.HealthCheck{
		Endpoint:       "/health",
		TimeoutSeconds: 5,
	})

	assert.Error(t, err, "Expected a network error")
}

// Additional test cases can be added here
