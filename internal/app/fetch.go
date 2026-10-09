package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/axelrindle/traefik2traefik/internal"
	dynamic "github.com/traefik/traefik/v3/pkg/config/dynamic"
	"go.uber.org/zap"
)

func strToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (a *App) fetchTraefikRouters() ([]internal.RouterRepresentation, error) {
	url := fmt.Sprintf("%s/api/http/routers?status=enabled", a.Config.Traefik.Endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		a.ready = false
		return nil, err
	}

	if a.Config.Traefik.BasicAuth != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", strToBase64(a.Config.Traefik.BasicAuth)))
	}

	res, err := a.http.Do(req)
	if err != nil {
		a.ready = false
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		a.ready = false
		return nil, err
	}

	data := []internal.RouterRepresentation{}
	json.Unmarshal(body, &data)

	return data, nil
}

// matchesRuleFilter reports whether rule passes the configured rule filters.
// Without filters every rule passes, otherwise at least one filter must match.
func (a *App) matchesRuleFilter(router internal.RouterRepresentation) bool {
	filters := a.Config.Traefik.RuleFilter
	if len(filters) == 0 {
		return true
	}

	for _, f := range filters {
		switch f.Type {
		case "rule:contains":
			if strings.Contains(router.Rule, f.Filter) {
				return true
			}
		case "rule:regex":
			re, err := regexp.Compile(f.Filter)
			if err != nil {
				a.Logger.Error("invalid regex in ruleFilter", zap.String("filter", f.Filter), zap.Error(err))
				continue
			}
			if re.MatchString(router.Rule) {
				return true
			}
		default:
			a.Logger.Warn("invalid filter type", zap.String("type", f.Type))
		}
	}

	return false
}

func (a *App) buildTraefikConfiguration(src []internal.RouterRepresentation) dynamic.Configuration {
	routers := map[string]*dynamic.Router{}
	services := map[string]*dynamic.Service{}

	servers := make([]dynamic.Server, len(a.Config.Generator.TargetServers))
	for i, server := range a.Config.Generator.TargetServers {
		servers[i].URL = server
	}

	for _, router := range src {
		if router.Status != "enabled" {
			continue
		}
		if !a.matchesRuleFilter(router) {
			continue
		}

		name, _, _ := strings.Cut(router.Name, "@")
		routers[name] = &dynamic.Router{
			Rule:        router.Rule,
			Service:     name,
			EntryPoints: a.Config.Generator.Entrypoints,
			Middlewares: a.Config.Generator.Middlewares,
			TLS:         a.Config.Generator.TLS,
		}
		services[name] = &dynamic.Service{
			LoadBalancer: &dynamic.ServersLoadBalancer{
				PassHostHeader: &a.Config.Generator.PassHostHeader,
				Servers:        servers,
			},
		}
	}

	configuration := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:  routers,
			Services: services,
		},
	}

	return configuration
}
