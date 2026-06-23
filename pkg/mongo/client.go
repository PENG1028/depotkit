package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect creates a new MongoDB client and verifies connectivity.
func Connect(host string, port int, database string) (*mongo.Client, *mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s:%d/%s", host, port, database)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("pinging MongoDB: %w", err)
	}

	db := client.Database(database)
	return client, db, nil
}

// Ping checks if MongoDB is reachable.
func Ping(host string, port int, database string) error {
	client, db, err := Connect(host, port, database)
	if err != nil {
		return err
	}
	_ = db
	ctx := context.Background()
	return client.Disconnect(ctx)
}

// ListCollections lists all collections in the database.
func ListCollections(db *mongo.Database) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("listing collections: %w", err)
	}

	return collections, nil
}
