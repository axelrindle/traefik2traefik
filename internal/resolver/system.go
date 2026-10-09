package resolver

import (
	"context"
	"errors"
	"net"

	probing "github.com/prometheus-community/pro-bing"
)

var errNoReply = errors.New("no echo reply received")

func systemLookup(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	ips := make([]net.IP, len(addrs))
	for i, addr := range addrs {
		ips[i] = addr.IP
	}
	return ips, nil
}

func pingProbe(ctx context.Context, ip net.IP) error {
	p, err := probing.NewPinger(ip.String())
	if err != nil {
		return err
	}
	p.Count = 1

	// RunWithContext returns nil on timeout as well, so the reply has to be counted.
	if err := p.RunWithContext(ctx); err != nil {
		return err
	}
	if p.Statistics().PacketsRecv == 0 {
		return errNoReply
	}
	return nil
}
