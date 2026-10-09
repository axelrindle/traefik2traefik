package app

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/axelrindle/traefik2traefik/internal/config"
	"github.com/axelrindle/traefik2traefik/internal/resolver"
	"github.com/go-co-op/gocron/v2"
	dynamic "github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type App struct {
	Config *config.Config
	Logger *zap.Logger

	// Resolver maps the Traefik endpoint host name to a reachable address.
	Resolver resolver.Resolver

	http      *http.Client
	scheduler gocron.Scheduler
	server    *http.Server

	ready         bool
	configuration dynamic.Configuration
}

func (a *App) IsReady() bool {
	return a.ready
}
