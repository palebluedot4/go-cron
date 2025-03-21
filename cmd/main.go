package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-cron/pkg/utils/timeutil"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health"
		},
	}))
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "ok",
			"timestamp": time.Now().In(timeutil.TaipeiLocation).Format(time.RFC3339Nano),
		})
	})

	go func() {
		if err := e.Start(":8080"); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	quiz := make(chan os.Signal, 1)
	signal.Notify(quiz, syscall.SIGINT, syscall.SIGTERM)
	<-quiz

	slog.Warn("Shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server", "error", err)
		os.Exit(1)
	}
}
