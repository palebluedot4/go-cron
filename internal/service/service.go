package service

import (
	"go-cron/internal/repository"
)

type Container struct{}

func New(repo *repository.Container) *Container {
	return &Container{}
}
