package utils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPortAvailable(t *testing.T) {
	// Initialize PortManager
	pm := NewPortManager(3000, 4000, 1)

	// Define a port known to be available for testing
	testPort := 3000

	// Check if the port is available
	assert.True(t, pm.IsPortAvailable(testPort), "Port should be available")
}

func TestAllocatePort(t *testing.T) {
	// Initialize PortManager
	pm := NewPortManager(3000, 4000, 1)

	// Attempt to allocate a port
	port, err := pm.AllocatePort()

	// Check if a port was successfully allocated and no error occurred
	assert.Greater(t, port, 0, "Allocated port should be greater than 0")
	assert.Nil(t, err, "No error should occur when allocating a port")
}

func TestNoAvailablePorts(t *testing.T) {
	// Initialize PortManager with a very limited range
	pm := NewPortManager(3000, 3000, 1)

	ln, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		t.Fatalf("Failed to manually occupy the port: %v", err)
	}
	defer ln.Close()

	// Attempt to allocate another port, which should fail
	_, err = pm.AllocatePort()
	assert.NotNil(t, err, "Error should occur when no ports are available")
}
