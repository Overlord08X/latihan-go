package database

import (
	"context"
	"fmt"
	"time"

	"api-students/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// NewMongoClient menyiapkan client MongoDB dan memastikan servernya benar-benar
// dapat dihubungi melalui Ping.
func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	uri := config.GetEnv("MONGO_URI", "mongodb://localhost:27017")
	opts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(5 * time.Second).
		SetMaxPoolSize(uint64(config.GetEnvInt("MONGO_MAX_POOL", 20)))

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("menyiapkan client MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return client, nil
}
