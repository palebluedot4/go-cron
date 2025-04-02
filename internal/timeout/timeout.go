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
