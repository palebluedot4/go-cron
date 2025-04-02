package controller

import (
	"go-cron/internal/service"
)

type Container struct{}

func New(svc *service.Container) *Container {
	return &Container{}
}
