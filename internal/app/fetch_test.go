package app

import (
	"testing"

	"github.com/axelrindle/traefik2traefik/internal"
	"github.com/axelrindle/traefik2traefik/internal/config"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/config/runtime"
	"go.uber.org/zap"
)

func router(name, rule string) internal.RouterRepresentation {
	return internal.RouterRepresentation{
		RouterInfo: &traefik.RouterInfo{
			Router: &dynamic.Router{Rule: rule},
			Status: "enabled",
		},
		Name: name,
	}
}

func TestBuildTraefikConfigurationRuleFilter(t *testing.T) {
	src := []internal.RouterRepresentation{
		router("a@kubernetes", "Host(`app.example.dev`)"),
		router("b@kubernetes", "Host(`other.local`)"),
		router("c@kubernetes", "Host(`foo`)"),
	}

	tests := []struct {
		name    string
		filters []config.FilterConfig
		want    []string
	}{
		{"no filter", nil, []string{"a", "b", "c"}},
		{"simple", []config.FilterConfig{{Type: "simple", Filter: "example.dev"}}, []string{"a"}},
		{"regex", []config.FilterConfig{{Type: "regex", Filter: "^Host\\(`foo`\\)$"}}, []string{"c"}},
		{"any matches", []config.FilterConfig{
			{Type: "simple", Filter: "example.dev"},
			{Type: "regex", Filter: "foo"},
		}, []string{"a", "c"}},
		{"invalid regex", []config.FilterConfig{{Type: "regex", Filter: "("}}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{Logger: zap.NewNop(), Config: &config.Config{}}
			a.Config.Traefik.RuleFilter = tt.filters

			got := a.buildTraefikConfiguration(src).HTTP.Routers
			if len(got) != len(tt.want) {
				t.Fatalf("got %d routers, want %d", len(got), len(tt.want))
			}
			for _, name := range tt.want {
				if _, ok := got[name]; !ok {
					t.Errorf("router %q missing", name)
				}
			}
		})
	}
}
