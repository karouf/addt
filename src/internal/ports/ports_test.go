package ports

import (
	"net"
	"testing"
)

func TestIsPortAvailable_Available(t *testing.T) {
	// Use a high port that's unlikely to be in use
	port := 59123

	if !IsPortAvailable(port) {
		t.Skip("Port 59123 is in use, skipping test")
	}

	if !IsPortAvailable(port) {
		t.Errorf("IsPortAvailable(%d) = false, want true", port)
	}
}

func TestIsPortAvailable_InUse(t *testing.T) {
	// Start a listener on a port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Get the port that was assigned
	port := listener.Addr().(*net.TCPAddr).Port

	if IsPortAvailable(port) {
		t.Errorf("IsPortAvailable(%d) = true, want false (port is in use)", port)
	}
}

func TestFindAvailablePort_FromStart(t *testing.T) {
	startPort := 59200

	port := FindAvailablePort(startPort)

	if port < startPort {
		t.Errorf("FindAvailablePort(%d) = %d, want >= %d", startPort, port, startPort)
	}

	// The returned port should be available
	if !IsPortAvailable(port) {
		t.Errorf("FindAvailablePort returned %d but it's not available", port)
	}
}

func TestFindAvailablePort_SkipsInUse(t *testing.T) {
	// Start a listener on a specific port
	listener, err := net.Listen("tcp", "localhost:59300")
	if err != nil {
		t.Skip("Could not bind to port 59300, skipping test")
	}
	defer listener.Close()

	// FindAvailablePort should skip the in-use port
	port := FindAvailablePort(59300)

	if port == 59300 {
		t.Errorf("FindAvailablePort(59300) = 59300, but that port is in use")
	}

	if port < 59300 {
		t.Errorf("FindAvailablePort(59300) = %d, want >= 59300", port)
	}
}
