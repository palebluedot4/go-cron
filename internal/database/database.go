package database

import (
	"context"
	"fmt"
	"sync"

	"go-cron/internal/config"
	"go-cron/internal/database/mongo"
	"go-cron/internal/database/postgres"
	"go-cron/internal/database/redis"
	"go-cron/pkg/logger"
)

type Database interface {
	IsConnected(ctx context.Context) bool
	Close(ctx context.Context) error
}

type Manager struct {
	cfg      *config.Config
	postgres *postgres.Postgres
	mongo    *mongo.MongoDB
	redis    *redis.Redis
	mu       sync.RWMutex
}

var log = logger.Instance()

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

func (m *Manager) Init(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	if m.cfg.Storage.Postgres.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pg, err := m.Postgres(ctx)
			if err != nil {
				log.WithError(err).Warn("Failed to initialize PostgreSQL connection")
				errCh <- fmt.Errorf("PostgreSQL connection failed: %w", err)
			}

			if pg != nil {
				log.WithField("database", "postgres").Info("Database connection established successfully")
			}
		}()
	}

	if m.cfg.Storage.MongoDB.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mdb, err := m.Mongo(ctx)
			if err != nil {
				log.WithError(err).Warn("Failed to initialize MongoDB connection")
				errCh <- fmt.Errorf("MongoDB connection failed: %w", err)
			}

			if mdb != nil {
				log.WithField("database", "mongodb").Info("Database connection established successfully")
			}
		}()
	}

	if m.cfg.Storage.Redis.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rdb, err := m.Redis(ctx)
			if err != nil {
				log.WithError(err).Warn("Failed to initialize Redis connection")
				errCh <- fmt.Errorf("Redis connection failed: %w", err)
			}

			if rdb != nil {
				log.WithField("database", "redis").Info("Database connection established successfully")
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("critical database initialization failed: %v", errors)
	}
	return nil
}

func (m *Manager) Postgres(ctx context.Context) (*postgres.Postgres, error) {
	if !m.cfg.Storage.Postgres.Enabled {
		return nil, nil
	}

	m.mu.RLock()
	if m.postgres != nil {
		defer m.mu.RUnlock()
		return m.postgres, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	pg, err := postgres.New(ctx, m.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres: %w", err)
	}

	m.postgres = pg
	return pg, nil
}

func (m *Manager) Mongo(ctx context.Context) (*mongo.MongoDB, error) {
	if !m.cfg.Storage.MongoDB.Enabled {
		return nil, nil
	}

	m.mu.RLock()
	if m.mongo != nil {
		defer m.mu.RUnlock()
		return m.mongo, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	mdb, err := mongo.New(ctx, m.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mongo: %w", err)
	}

	m.mongo = mdb
	return mdb, nil
}

func (m *Manager) Redis(ctx context.Context) (*redis.Redis, error) {
	if !m.cfg.Storage.Redis.Enabled {
		return nil, nil
	}

	m.mu.RLock()
	if m.redis != nil {
		defer m.mu.RUnlock()
		return m.redis, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	rdb, err := redis.New(ctx, m.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	m.redis = rdb
	return rdb, nil
}

func (m *Manager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	if m.postgres != nil {
		if err := m.postgres.Close(ctx); err != nil {
			log.WithError(err).Error("Error closing PostgreSQL connection")
			errs = append(errs, fmt.Errorf("PostgreSQL close error: %w", err))
		} else {
			log.WithField("database", "postgres").Info("Database connection closed successfully")
		}
		m.postgres = nil
	}

	if m.mongo != nil {
		if err := m.mongo.Close(ctx); err != nil {
			log.WithError(err).Error("Error closing MongoDB connection")
			errs = append(errs, fmt.Errorf("MongoDB close error: %w", err))
		} else {
			log.WithField("database", "mongodb").Info("Database connection closed successfully")
		}
		m.mongo = nil
	}

	if m.redis != nil {
		if err := m.redis.Close(ctx); err != nil {
			log.WithError(err).Error("Error closing Redis connection")
			errs = append(errs, fmt.Errorf("Redis close error: %w", err))
		} else {
			log.WithField("database", "redis").Info("Database connection closed successfully")
		}
		m.redis = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}
	return nil
}
