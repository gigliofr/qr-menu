package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache wraps Redis client for caching
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Test connessione
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
	}, nil
}

func (rc *RedisCache) operationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

// Get retrieves a value from Redis cache
func (rc *RedisCache) Get(key string) (string, error) {
	ctx, cancel := rc.operationContext()
	defer cancel()
	return rc.client.Get(ctx, key).Result()
}

// Set stores a value in Redis cache with TTL
func (rc *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	ctx, cancel := rc.operationContext()
	defer cancel()
	return rc.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a value from Redis cache
func (rc *RedisCache) Delete(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	ctx, cancel := rc.operationContext()
	defer cancel()
	return rc.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in Redis
func (rc *RedisCache) Exists(key string) bool {
	ctx, cancel := rc.operationContext()
	defer cancel()
	return rc.client.Exists(ctx, key).Val() > 0
}

// FlushAll deletes all keys from Redis
func (rc *RedisCache) FlushAll() error {
	ctx, cancel := rc.operationContext()
	defer cancel()
	return rc.client.FlushAll(ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// Ping tests Redis connectivity
func (rc *RedisCache) Ping() error {
	ctx, cancel := rc.operationContext()
	defer cancel()
	return rc.client.Ping(ctx).Err()
}

// Info returns Redis server info (useful for health checks)
func (rc *RedisCache) Info() (string, error) {
	ctx, cancel := rc.operationContext()
	defer cancel()
	return rc.client.Info(ctx).Val(), nil
}
