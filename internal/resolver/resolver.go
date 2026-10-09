// Package resolver turns host names into IP addresses that are known to be reachable.
package resolver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

// ErrNoReachable is returned when a host resolves to no address or none of its addresses answer.
var ErrNoReachable = errors.New("could not resolve a reachable address")

// Resolver returns a reachable IP address for a host name.
type Resolver interface {
	Resolve(ctx context.Context, host string) (net.IP, error)
}

// LookupFunc returns all addresses of a host, usually by asking DNS.
type LookupFunc func(ctx context.Context, host string) ([]net.IP, error)

// ProbeFunc returns nil if the address is reachable.
type ProbeFunc func(ctx context.Context, ip net.IP) error

const probeTimeout = 5 * time.Second

type resolver struct {
	lookup LookupFunc
	probe  ProbeFunc
}

// New combines a lookup and a probe. The first address, in lookup order, that passes the probe wins.
func New(lookup LookupFunc, probe ProbeFunc) Resolver {
	return &resolver{lookup: lookup, probe: probe}
}

// NewResolver resolves via the system resolver and checks reachability with an ICMP echo.
func NewResolver() Resolver {
	return New(systemLookup, pingProbe)
}

func (r *resolver) Resolve(ctx context.Context, host string) (net.IP, error) {
	ips, err := r.lookup(ctx, host)
	if err != nil {
		return nil, err
	}

	var errs []error
	for _, ip := range ips {
		err := r.probeOnce(ctx, ip)
		if err == nil {
			return ip, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", ip, err))
	}

	return nil, fmt.Errorf("%w for %q: %w", ErrNoReachable, host, errors.Join(errs...))
}

func (r *resolver) probeOnce(ctx context.Context, ip net.IP) error {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	return r.probe(ctx, ip)
}
