package timeout

import (
	"time"

	"go-cron/internal/config"
)

func ServerShutdown(cfg *config.Config) time.Duration {
	if cfg.Server.ShutdownTimeout == 0 {
		return 60 * time.Second
	}
	return cfg.Server.ShutdownTimeout
}

func ServerRead(cfg *config.Config) time.Duration {
	if cfg.Server.ReadTimeout == 0 {
		return 15 * time.Second
	}
	return cfg.Server.ReadTimeout
}

func ServerWrite(cfg *config.Config) time.Duration {
	if cfg.Server.WriteTimeout == 0 {
		return 30 * time.Second
	}
	return cfg.Server.WriteTimeout
}

func ServerIdle(cfg *config.Config) time.Duration {
	if cfg.Server.IdleTimeout == 0 {
		return 120 * time.Second
	}
	return cfg.Server.IdleTimeout
}

func DatabaseConnect(cfg *config.Config) time.Duration {
	if cfg.Storage.ConnectTimeout == 0 {
		return 15 * time.Second
	}
	return cfg.Storage.ConnectTimeout
}

func DatabaseShutdown(cfg *config.Config) time.Duration {
	if cfg.Storage.ShutdownTimeout == 0 {
		return 30 * time.Second
	}
	return cfg.Storage.ShutdownTimeout
}
