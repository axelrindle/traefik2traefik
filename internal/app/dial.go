package app

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// newHTTPClient builds the client for talking to the source Traefik.
// Host names are resolved through a.Resolver, but the request URL stays untouched,
// so Host header and TLS server name still carry the original name.
func (a *App) newHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = a.dialContext
	transport.DisableKeepAlives = true
	if a.Config.Traefik.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &http.Client{
		Timeout:   time.Duration(a.Config.Traefik.Timeout) * time.Second,
		Transport: transport,
	}
}

func (a *App) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if net.ParseIP(host) == nil {
		ip, err := a.Resolver.Resolve(ctx, host)
		if err != nil {
			return nil, err
		}

		ips := ip.String()
		a.Logger.Debug("resolved host", zap.String("host", host), zap.String("ip", ips))
		addr = net.JoinHostPort(ips, port)
	}

	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	return dialer.DialContext(ctx, network, addr)
}
