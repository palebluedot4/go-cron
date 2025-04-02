package postgres

import (
	"context"
	"fmt"

	"go-cron/internal/config"
	"go-cron/internal/environment"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Postgres struct {
	Gorm *gorm.DB
}

func New(ctx context.Context, cfg *config.Config) (*Postgres, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s application_name=go-cron",
		cfg.Storage.Postgres.Host,
		cfg.Storage.Postgres.Port,
		cfg.Storage.Postgres.User,
		cfg.Storage.Postgres.Password,
		cfg.Storage.Postgres.DBName,
		cfg.Storage.Postgres.SSLMode,
	)

	logLevel := gormLogLevelForEnv(cfg.Server.Env)

	dbConfig := &gorm.Config{
		PrepareStmt: false,
		Logger:      gormLogger.Default.LogMode(logLevel),
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), dbConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Storage.Postgres.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Storage.Postgres.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Storage.Postgres.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Storage.Postgres.ConnMaxIdleTime)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	return &Postgres{
		Gorm: db,
	}, nil
}

func (p *Postgres) Close(ctx context.Context) error {
	sqlDB, err := p.Gorm.DB()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return sqlDB.Close()
	}
}

func (p *Postgres) IsConnected(ctx context.Context) bool {
	sqlDB, err := p.Gorm.DB()
	if err != nil {
		return false
	}

	return sqlDB.PingContext(ctx) == nil
}

func gormLogLevelForEnv(env environment.Environment) gormLogger.LogLevel {
	if environment.IsProduction(env) {
		return gormLogger.Error
	} else if environment.IsStaging(env) {
		return gormLogger.Warn
	}
	return gormLogger.Info
}
