package resolver

import (
	"context"
	"fmt"
	"net"
)

// NewTestResolver answers from a static host table and needs no network.
// Addresses listed in unreachable fail the probe, all others pass.
func NewTestResolver(hosts map[string][]net.IP, unreachable ...net.IP) Resolver {
	lookup := func(_ context.Context, host string) ([]net.IP, error) {
		ips, ok := hosts[host]
		if !ok {
			return nil, &net.DNSError{Err: "no such host", Name: host, IsNotFound: true}
		}
		return ips, nil
	}

	probe := func(_ context.Context, ip net.IP) error {
		for _, down := range unreachable {
			if down.Equal(ip) {
				return fmt.Errorf("%s is unreachable", ip)
			}
		}
		return nil
	}

	return New(lookup, probe)
}
