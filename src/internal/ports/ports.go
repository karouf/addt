package ports

import (
	"fmt"
	"net"
	"time"
)

// IsPortAvailable checks if a port is available for binding
func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

// FindAvailablePort finds the next available port starting from startPort
func FindAvailablePort(startPort int) int {
	port := startPort
	for !IsPortAvailable(port) {
		port++
	}
	return port
}
