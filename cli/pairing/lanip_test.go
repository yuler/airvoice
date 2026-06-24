package pairing

import (
	"net"
	"testing"
)

func TestLocalIPv4(t *testing.T) {
	ipStr, err := LocalIPv4()
	if err != nil {
		// If there is an error (e.g. offline environment), we make sure it is handled gracefully
		// and the string returned is empty.
		t.Logf("LocalIPv4 returned error (acceptable if offline): %v", err)
		if ipStr != "" {
			t.Errorf("expected empty string when error is returned, got %q", ipStr)
		}
		return
	}

	if ipStr == "" {
		t.Error("expected non-empty IP string when error is nil")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		t.Errorf("returned string %q is not a valid IP address", ipStr)
	}

	if ip.To4() == nil {
		t.Errorf("expected IPv4 address, got %q", ipStr)
	}

	if ip.IsLoopback() {
		t.Errorf("expected non-loopback IP address, got loopback %q", ipStr)
	}
}
