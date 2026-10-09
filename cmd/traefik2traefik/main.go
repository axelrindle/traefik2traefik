package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"github.com/axelrindle/traefik2traefik/internal/app"
	"github.com/axelrindle/traefik2traefik/internal/config"
	"github.com/axelrindle/traefik2traefik/internal/resolver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed banner.txt
var banner string

var (
	Version        = "dev"
	CommitHash     = "unknown"
	BuildTimestamp = time.Now().Format("2006-01-02T15:04:05")
)

var (
	configFile  string
	showVersion bool
	healthcheck bool
)

func BuildVersion() string {
	return fmt.Sprintf("%s-%s (%s)", Version, CommitHash, BuildTimestamp)
}

func makeLogger(c *config.Config) (*zap.Logger, error) {
	if c.Environment == "production" {
		conf := zap.NewProductionConfig()
		conf.Level = c.ZapLoggerLevel()
		conf.DisableCaller = true
		return conf.Build()
	} else {
		conf := zap.NewDevelopmentConfig()
		conf.Level = c.ZapLoggerLevel()
		conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return conf.Build()
	}
}

func main() {
	println(fmt.Sprintf(banner, BuildVersion()))

	flag.StringVar(&configFile, "config", "config.yml", "path to the config file")
	flag.BoolVar(&showVersion, "version", false, "show program version")
	flag.BoolVar(&healthcheck, "healthcheck", false, "run a healthcheck")
	flag.Parse()
	if showVersion {
		return
	}

	cfg := config.Load(configFile)

	logger, err := makeLogger(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	if healthcheck {
		addressParts := strings.Split(cfg.Server.Address, ":")
		if len(addressParts) != 2 {
			logger.Fatal("invalid server address", zap.String("address", cfg.Server.Address))
		}
		port := addressParts[1]

		res, err := http.Get("http://127.0.0.1:" + port + "/status")
		if err != nil {
			log.Fatal("healthcheck failed", zap.Error(err))
		}

		if res.StatusCode != 200 {
			log.Fatal("healthcheck failed", zap.Int("status", res.StatusCode))
		}

		return
	}

	logger.Info("starting program", zap.String("mode", cfg.Environment))

	app := &app.App{
		Config:   cfg,
		Logger:   logger.With(zap.String("component", "app")),
		Resolver: resolver.NewResolver(),
	}
	app.Init()

	go app.Boot()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutting down …")

	app.Shutdown()
}
