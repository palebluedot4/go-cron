package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go-cron/internal/config"
	"go-cron/internal/environment"
	"go-cron/pkg/logger"
	"go-cron/pkg/utils/timeutil"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	log := logger.Instance()
	defer logger.Shutdown()

	if err := config.Init(); err != nil {
		log.WithError(err).Fatal("Failed to initialize config")
	}

	cfg, err := config.Instance()
	if err != nil {
		log.WithError(err).Fatal("Failed to get config")
	}

	env := cfg.Server.Env
	if !environment.IsValid(env) {
		log.Fatal("Invalid environment configuration")
	}

	log.WithFields(map[string]any{
		"environment": env,
		"port":        cfg.Server.Port,
	}).Info("Starting application")

	e := echo.New()
	updateServerSettings(cfg, e)

	config.RegisterChangeCallback(func(cfg *config.Config, fileName string) {
		log.WithField("file", fileName).Info("Config file changed")
		updateLogSettings(cfg, log)
	})

	go func() {
		time.Sleep(10 * time.Second)
		log.AdjustForProduction()
	}()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health"
		},
	}))
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "ok",
			"timestamp": timeutil.Now().Format(time.RFC3339Nano),
		})
	})

	quit := make(chan os.Signal, 1)
	go func() {
		err := e.Start(":" + strconv.Itoa(cfg.Server.Port))
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("Server failed unexpectedly")
			quit <- syscall.SIGTERM
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Warn("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout(cfg))
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Failed to shutdown server")
		os.Exit(1)
	}

	log.Info("Server gracefully stopped")
}

func updateLogSettings(cfg *config.Config, log *logger.Logger) {
	if cfg.Server.LogLevel != "" {
		logger.SetLogLevel(cfg.Server.LogLevel)
	}

	if environment.IsProduction(cfg.Server.Env) {
		log.AdjustForProduction()
	}
}

func updateServerSettings(cfg *config.Config, e *echo.Echo) {
	if environment.IsDevelopment(cfg.Server.Env) {
		e.Debug = true
	} else {
		e.Debug = false
	}
}

func shutdownTimeout(cfg *config.Config) time.Duration {
	if cfg.Server.Timeout == 0 {
		return 10 * time.Second
	}
	return cfg.Server.Timeout
}
