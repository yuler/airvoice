package pairing

import (
	"errors"
	"net"
)

// LocalIPv4 walks the system's network interfaces and returns the first non-loopback IPv4 address as a string.
// If no active IPv4 interface is found, it returns an empty string and an error.
func LocalIPv4() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		// Check if interface is up
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ipv4 := ip.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}

	return "", errors.New("no active non-loopback IPv4 interface found")
}
