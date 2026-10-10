package config

import (
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type LoggingConfig struct {
	Level string `key:"level" default:"info" validate:"oneof=debug info warn error"`
}

type ServerConfig struct {
	Address string `key:"address" default:":1337"`
}

type FilterConfig struct {
	Type   string `key:"type" validate:"required"`
	Filter string `key:"filter" validate:"required"`
}

type TraefikConfig struct {
	Endpoint           string         `key:"endpoint" validate:"required,http_url"`
	BasicAuth          string         `key:"basicAuth"`
	RuleFilter         []FilterConfig `key:"ruleFilter" validate:"dive"`
	Timeout            time.Duration  `key:"timeout" default:"5s"`
	InsecureSkipVerify bool           `key:"insecureSkipVerify" default:"false"`
}

type GeneratorConfig struct {
	Entrypoints    []string                   `key:"entrypoints" validate:"required"`
	TargetServers  []string                   `key:"targets" validate:"required"`
	Middlewares    []string                   `key:"middlewares"`
	PassHostHeader bool                       `key:"passHostHeader" validate:"boolean" default:"false"`
	TLS            *dynamic.RouterTLSConfig   `key:"tls,omitempty"`
	HealthCheck    *dynamic.ServerHealthCheck `json:"healthCheck,omitempty"`
}

type Config struct {
	Logging   LoggingConfig   `default:""`
	Server    ServerConfig    `default:""`
	Traefik   TraefikConfig   `default:""`
	Generator GeneratorConfig `default:""`

	Environment     string        `key:"env" default:"production" validate:"oneof=production development"`
	RefreshInterval time.Duration `key:"refresh" default:"30s"`
}
