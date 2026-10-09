package resolver

import (
	"context"
	"errors"
	"net"
	"testing"
)

var (
	a = net.ParseIP("10.0.0.1")
	b = net.ParseIP("10.0.0.2")
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name        string
		hosts       map[string][]net.IP
		unreachable []net.IP
		host        string
		want        net.IP
		wantErr     error
	}{
		{"first reachable", map[string][]net.IP{"h": {a, b}}, nil, "h", a, nil},
		{"skips unreachable", map[string][]net.IP{"h": {a, b}}, []net.IP{a}, "h", b, nil},
		{"all unreachable", map[string][]net.IP{"h": {a, b}}, []net.IP{a, b}, "h", nil, ErrNoReachable},
		{"empty lookup result", map[string][]net.IP{"h": {}}, nil, "h", nil, ErrNoReachable},
		{"unknown host", map[string][]net.IP{}, nil, "h", nil, &net.DNSError{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTestResolver(tt.hosts, tt.unreachable...)

			got, err := r.Resolve(context.Background(), tt.host)

			var dnsErr *net.DNSError
			switch {
			case tt.wantErr == nil && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case errors.As(tt.wantErr, &dnsErr):
				if !errors.As(err, &dnsErr) {
					t.Fatalf("got %v, want a DNS error", err)
				}
			case tt.wantErr != nil && !errors.Is(err, tt.wantErr):
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}

			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveProbesInLookupOrder(t *testing.T) {
	var probed []string
	r := New(
		func(context.Context, string) ([]net.IP, error) { return []net.IP{a, b}, nil },
		func(_ context.Context, ip net.IP) error {
			probed = append(probed, ip.String())
			return nil
		},
	)

	if _, err := r.Resolve(context.Background(), "h"); err != nil {
		t.Fatal(err)
	}
	if len(probed) != 1 || probed[0] != a.String() {
		t.Errorf("probed %v, want only %v", probed, a)
	}
}
