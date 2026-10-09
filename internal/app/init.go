package app

import (
	"net"
	"net/http"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/goccy/go-yaml"
	"go.uber.org/zap"
)

func (a *App) Init() {
	a.http = a.newHTTPClient()

	a.initScheduler()
	a.initHttpServer()
}

func (a *App) Boot() {
	l, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		a.Logger.Fatal("failed to start http server", zap.Error(err))
	}

	a.Logger.Info("http server listening", zap.String("address", a.server.Addr))
	err = a.server.Serve(l)
	if err != http.ErrServerClosed {
		a.Logger.Fatal("failed to shutdown the http server", zap.Error(err))
	}
}

func (a *App) Shutdown() {
	var err error

	err = a.server.Close()
	if err != nil {
		a.Logger.Fatal("failed to shutdown the http server", zap.Error(err))
	}

	err = a.scheduler.Shutdown()
	if err != nil {
		a.Logger.Fatal("failed to shutdown the scheduler", zap.Error(err))
	}
}

func (a *App) initScheduler() {
	scheduler, err := gocron.NewScheduler(gocron.WithStopTimeout(time.Second * 1))
	if err != nil {
		a.Logger.Fatal("failed to create a scheduler", zap.Error(err))
	}
	a.scheduler = scheduler

	a.scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()),
		gocron.NewTask(a.sync),
	)
	a.scheduler.NewJob(
		gocron.DurationJob(time.Second*time.Duration(a.Config.RefreshInterval)),
		gocron.NewTask(a.sync),
	)
	a.scheduler.Start()
	a.Logger.Info("scheduler started", zap.Int64("interval", a.Config.RefreshInterval))
}

func (a *App) initHttpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/traefik", func(w http.ResponseWriter, r *http.Request) {
		if !a.ready {
			w.WriteHeader(503)
			return
		}

		if len(a.configuration.HTTP.Routers) == 0 {
			w.Header().Add("Content-Type", "text/yaml")
			w.WriteHeader(http.StatusOK)
			return
		}

		file, err := yaml.Marshal(a.configuration)
		if err != nil {
			a.Logger.Fatal("failed to generate yaml", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "text/yaml")
			w.WriteHeader(http.StatusOK)
			w.Write(file)
		}
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if a.ready {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	a.server = &http.Server{
		Addr:    a.Config.Server.Address,
		Handler: mux,
	}
}
