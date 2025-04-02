package mongo

import (
	"context"

	"go-cron/internal/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func New(ctx context.Context, cfg *config.Config) (*MongoDB, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.Storage.MongoDB.URI).SetMaxPoolSize(cfg.Storage.MongoDB.MaxPoolSize))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, err
	}

	return &MongoDB{
		Client:   client,
		Database: client.Database(cfg.Storage.MongoDB.Database),
	}, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) IsConnected(ctx context.Context) bool {
	return m.Client.Ping(ctx, nil) == nil
}
