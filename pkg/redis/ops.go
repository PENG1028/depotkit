package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// ScanKeys scans Redis for keys matching a pattern.
func ScanKeys(client *redis.Client, pattern string, count int64) ([]string, error) {
	ctx := context.Background()

	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error
		batch, cursor, err = client.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, fmt.Errorf("scanning keys with pattern %s: %w", pattern, err)
		}
		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// TTLReport checks TTL for a list of keys and returns those without TTL.
type TTLResult struct {
	Key        string
	TTL        int64 // seconds, -1 for no expiry, -2 for not found
	HasTTL     bool
}

// CheckTTL checks TTL status for all matching keys.
func CheckTTL(client *redis.Client, pattern string) ([]TTLResult, error) {
	keys, err := ScanKeys(client, pattern, 100)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var results []TTLResult

	for _, key := range keys {
		ttl, err := client.TTL(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("checking TTL for key %s: %w", key, err)
		}

		hasTTL := ttl > 0
		results = append(results, TTLResult{
			Key:    key,
			TTL:    int64(ttl.Seconds()),
			HasTTL: hasTTL,
		})
	}

	return results, nil
}

// FlushNamespace deletes all keys matching a pattern.
func FlushNamespace(client *redis.Client, pattern string) (int64, error) {
	keys, err := ScanKeys(client, pattern, 100)
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		return 0, nil
	}

	ctx := context.Background()
	count, err := client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("deleting keys: %w", err)
	}

	return count, nil
}
