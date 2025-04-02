package repository

import (
	"go-cron/internal/database"
)

type Container struct{}

func New(db *database.Manager) *Container {
	return &Container{}
}
