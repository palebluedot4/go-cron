package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go-cron/internal/config"
	"go-cron/internal/environment"
	"go-cron/pkg/utils/timeutil"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := config.Init(); err != nil {
		slog.Error("failed to initialize config", "error", err)
		os.Exit(1)
	}

	cfg, err := config.Instance()
	if err != nil {
		slog.Error("failed to get config", "error", err)
		os.Exit(1)
	}

	env := cfg.Server.Env
	if !environment.IsValid(env) {
		slog.Error("invalid environment configuration", "environment", env)
		os.Exit(1)
	}

	slog.Info("starting application", "environment", env, "port", cfg.Server.Port)

	e := echo.New()
	updateServerSettings(cfg, e)

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

	go func() {
		err := e.Start(":" + strconv.Itoa(cfg.Server.Port))
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server failed unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Warn("shutting down the server")

	shutdownTimeout := cfg.Server.Timeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	}

	slog.Error("failed to shutdown server", "error", err)
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
