package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Connect creates a new Redis client and verifies connectivity.
func Connect(host string, port int) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("connecting to Redis at %s: %w", addr, err)
	}

	return client, nil
}

// Ping checks if Redis is reachable.
func Ping(host string, port int) (string, error) {
	client, err := Connect(host, port)
	if err != nil {
		return "", err
	}
	defer client.Close()

	ctx := context.Background()
	return client.Ping(ctx).Result()
}
