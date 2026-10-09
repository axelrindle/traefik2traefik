package app

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/axelrindle/traefik2traefik/internal/config"
	"github.com/axelrindle/traefik2traefik/internal/resolver"
	"go.uber.org/zap"
)

func newDialTestApp(r resolver.Resolver) *App {
	return &App{Logger: zap.NewNop(), Config: &config.Config{}, Resolver: r}
}

func TestHTTPClientResolvesHostName(t *testing.T) {
	var gotHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	loopback := net.ParseIP("127.0.0.1")
	a := newDialTestApp(resolver.NewTestResolver(map[string][]net.IP{"traefik.test": {loopback}}))

	res, err := a.newHTTPClient().Get("http://traefik.test:" + u.Port() + "/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if want := "traefik.test:" + u.Port(); gotHost != want {
		t.Errorf("server saw host %q, want %q", gotHost, want)
	}
}

func TestHTTPClientFailsWithoutReachableAddress(t *testing.T) {
	loopback := net.ParseIP("127.0.0.1")
	a := newDialTestApp(resolver.NewTestResolver(
		map[string][]net.IP{"traefik.test": {loopback}},
		loopback,
	))

	if _, err := a.newHTTPClient().Get("http://traefik.test:1/"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestHTTPClientSkipsResolverForIPLiterals(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()

	// empty host table: any lookup would fail
	a := newDialTestApp(resolver.NewTestResolver(nil))

	res, err := a.newHTTPClient().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
}

func TestHTTPClientResolvesOnEveryRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	lookups := 0
	a := newDialTestApp(resolver.New(
		func(context.Context, string) ([]net.IP, error) {
			lookups++
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
		func(context.Context, net.IP) error { return nil },
	))
	client := a.newHTTPClient()

	for i := 0; i < 3; i++ {
		res, err := client.Get("http://traefik.test:" + u.Port() + "/")
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
	}

	if lookups != 3 {
		t.Errorf("got %d lookups, want 3", lookups)
	}
}
