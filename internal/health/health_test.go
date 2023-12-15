package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/stretchr/testify/assert"
)

type sleepCounterClock struct {
	clockwork.FakeClock
	sleepCount int
}

func (s *sleepCounterClock) Sleep(d time.Duration) {
	s.sleepCount++
	s.FakeClock.Sleep(d)
}

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
		Endpoint:        "/",
		TimeoutSeconds:  5,
		IntervalSeconds: 2,
		MaxRetries:      3,
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

func TestCheckHealth_SleepWithError(t *testing.T) {
	t.Parallel()

	clk := &sleepCounterClock{FakeClock: clockwork.NewFakeClock()}

	serverURL, teardown := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // Return 500 Internal Server Error
	})

	defer teardown()

	go func() {
		for {
			clk.Advance(1 * time.Second)
		}
	}()

	err := CheckHealthWithClock(serverURL, &config.HealthCheck{
		Endpoint:        "/health",
		TimeoutSeconds:  5,
		IntervalSeconds: 2,
		MaxRetries:      3,
	}, clk)

	assert.Error(t, err, "Expected an error for server error response")
	assert.Greater(t, clk.sleepCount, 0, "Expected Sleep to be called at least once")
}

func TestCheckHealthWithClock_StalledServer(t *testing.T) {
	t.Parallel()

	// Create a test server that delays its response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Delay response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clk := clockwork.NewFakeClock()

	// Run CheckHealthWithClock in a separate goroutine, as it will block due to the server delay
	go func() {
		err := CheckHealthWithClock(server.URL, &config.HealthCheck{
			Endpoint:        "/health",
			TimeoutSeconds:  2, // This is less than the server delay
			IntervalSeconds: 2,
			MaxRetries:      1,
		}, clk)

		// We expect an error, as the server response will be delayed beyond the timeout
		assert.Error(t, err, "Expected an error due to server response delay")
	}()

	// Advance the clock in a loop to simulate time passing
	for i := 0; i < 5; i++ {
		clk.Advance(1 * time.Second)
		time.Sleep(1 * time.Second) // This is needed to allow the CheckHealthWithClock goroutine to progress
	}
}
