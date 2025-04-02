package server

import (
	"strconv"

	"go-cron/internal/config"
	"go-cron/internal/environment"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Address(cfg *config.Config) string {
	if cfg.Server.Port == 0 {
		return ":8080"
	}
	return ":" + strconv.Itoa(cfg.Server.Port)
}

func Configure(e *echo.Echo, cfg *config.Config) {
	if environment.IsDevelopment(cfg.Server.Env) {
		e.Debug = true
	} else {
		e.Debug = false
	}
}

func SetupMiddleware(e *echo.Echo, cfg *config.Config) {
	e.Pre(middleware.RemoveTrailingSlash())

	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health"
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
}
