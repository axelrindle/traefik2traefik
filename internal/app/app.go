package app

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/axelrindle/traefik2traefik/internal/config"
	"github.com/go-co-op/gocron/v2"
	dynamic "github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type App struct {
	Config *config.Config
	Logger *zap.Logger

	http      *http.Client
	scheduler gocron.Scheduler
	server    *http.Server

	ready         bool
	configuration dynamic.Configuration
}

func (a *App) IsReady() bool {
	return a.ready
}
