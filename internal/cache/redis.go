package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ CacheClient = (*RedisClient)(nil)

type RedisClient struct {
	client *redis.Client
}

func (c *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	bs, err := c.client.Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	return bs, err
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	res := c.client.Set(ctx, key, value, expiration)
	return res.Err()
}

func (c *RedisClient) Close() error {
	return c.client.Close()
}
