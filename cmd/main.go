package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-cron/internal/config"
	"go-cron/internal/database"
	"go-cron/internal/environment"
	"go-cron/internal/server"
	"go-cron/internal/timeout"
	"go-cron/pkg/logger"
	"go-cron/pkg/utils/timeutil"

	"github.com/labstack/echo/v4"
)

func main() {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

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

	dbManager := database.NewManager(cfg)
	dbInitCtx, dbInitCancel := context.WithTimeout(rootCtx, timeout.DatabaseConnect(cfg))
	defer dbInitCancel()
	if err := dbManager.Init(dbInitCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.WithField("timeout", timeout.DatabaseConnect(cfg)).Fatal("Database initialization timed out")
		}
		log.WithError(err).Fatal("Failed to initialize database")
	}

	defer func() {
		dbCloseCtx, dbCloseCancel := context.WithTimeout(context.Background(), timeout.DatabaseShutdown(cfg))
		defer dbCloseCancel()

		if err := dbManager.Close(dbCloseCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.WithField("timeout", timeout.DatabaseShutdown(cfg)).Error("Database closing timed out")
			} else {
				log.WithError(err).Error("Failed to close database connections")
			}
		}
	}()

	e := echo.New()
	server.Configure(e, cfg)
	server.SetupMiddleware(e, cfg)

	config.RegisterChangeCallback(func(cfg *config.Config, fileName string) {
		log.WithField("file", fileName).Info("Config file changed")
		logger.UpdateFromConfig(cfg)
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "ok",
			"timestamp": timeutil.Now().Format(time.RFC3339Nano),
		})
	})

	s := &http.Server{
		Addr:         server.Address(cfg),
		ReadTimeout:  timeout.ServerRead(cfg),
		WriteTimeout: timeout.ServerWrite(cfg),
		IdleTimeout:  timeout.ServerIdle(cfg),
	}
	e.Server = s

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := e.StartServer(s)
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("Server failed unexpectedly")
			quit <- syscall.SIGTERM
		}
	}()

	<-quit
	rootCancel()

	log.Info("Shutting down the server")

	serverShutdownCtx, serverShutdownCancel := context.WithTimeout(context.Background(), timeout.ServerShutdown(cfg))
	defer serverShutdownCancel()

	if err := e.Shutdown(serverShutdownCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.WithField("timeout", timeout.ServerShutdown(cfg)).Error("Server shutdown timed out")
		} else {
			log.WithError(err).Error("Failed to shutdown server")
		}
		os.Exit(1)
	}

	log.Info("Server gracefully stopped")
}
