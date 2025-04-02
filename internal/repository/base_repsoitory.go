package repository

import (
	"context"

	"go-cron/internal/database"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

type BaseRepository struct {
	db *database.Manager
}

func (r *BaseRepository) Postgres(ctx context.Context) (*gorm.DB, error) {
	pg, err := r.db.Postgres(ctx)
	if err != nil {
		return nil, err
	}

	return pg.Gorm, nil
}

func (r *BaseRepository) Redis(ctx context.Context) (*redis.Client, error) {
	rdb, err := r.db.Redis(ctx)
	if err != nil {
		return nil, err
	}

	return rdb.Client, nil
}

func (r *BaseRepository) MongoDB(ctx context.Context) (*mongo.Database, error) {
	mdb, err := r.db.MongoDB(ctx)
	if err != nil {
		return nil, err
	}

	return mdb.Database, nil
}
