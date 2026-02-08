package orbstack

import (
	"net"
	"testing"
)

func TestGetHostGatewayIP(t *testing.T) {
	ip, err := getHostGatewayIP()
	if err != nil {
		t.Skipf("Skipping: no network available to detect host IP: %v", err)
	}

	if ip == "" {
		t.Fatal("expected non-empty IP address")
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		t.Fatalf("returned value %q is not a valid IP address", ip)
	}

	if parsed.IsLoopback() {
		t.Fatalf("expected non-loopback IP, got %s", ip)
	}
}
