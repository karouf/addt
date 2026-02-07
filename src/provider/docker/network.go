package docker

import (
	"fmt"
	"net"
)

// getHostGatewayIP detects the host's IP address that is reachable from containers.
// On macOS, Docker Desktop runs containers inside a Linux VM. This function detects
// the host's outbound IP via the routing table for TCP-based proxy connections.
func getHostGatewayIP() (string, error) {
	// Use a UDP dial to determine the outbound IP without sending any traffic.
	// The OS routing table resolves which local IP would be used to reach 8.8.8.8.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("failed to detect host IP: %w", err)
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", fmt.Errorf("unexpected address type from UDP dial")
	}

	return localAddr.IP.String(), nil
}
