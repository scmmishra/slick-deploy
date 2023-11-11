package utils

import (
	"fmt"
	"net"
	"sync"
)

// PortManager manages the allocation and deallocation of network ports.
type PortManager struct {
	StartPort     int
	MaxPort       int
	Allocated     map[int]bool
	PortIncrement int
	mu            sync.Mutex
}

// NewPortManager creates a new instance of PortManager.
func NewPortManager(startPort, maxPort, portIncrement int) *PortManager {
	return &PortManager{
		StartPort:     startPort,
		MaxPort:       maxPort,
		PortIncrement: portIncrement,
	}
}

// IsPortAvailable checks if a port is available for use.
func (pm *PortManager) IsPortAvailable(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}

	ln.Close()
	return true
}

// AllocatePort finds and allocates an available port.
func (pm *PortManager) AllocatePort() (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for port := pm.StartPort; port <= pm.MaxPort; port += pm.PortIncrement {
		if pm.IsPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports")
}
