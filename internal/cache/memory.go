package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
)

var _ CacheClient = (*InMemoryCacheClient)(nil)

type InMemoryCacheClient struct {
	storage *cache.Cache
}

func NewInMemoryCacheClient() *InMemoryCacheClient {
	return &InMemoryCacheClient{
		storage: cache.New(
			time.Duration(5*time.Minute),
			time.Duration(10*time.Minute),
		),
	}
}

func (c *InMemoryCacheClient) Get(_ context.Context, key string) ([]byte, error) {
	x, ok := c.storage.Get(key)
	if !ok {
		return nil, ErrNotFound
	}
	val, ok := x.([]byte)
	if !ok {
		return nil, ErrInvalidData
	}
	return val, nil
}

func (c *InMemoryCacheClient) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	c.storage.Set(key, value, expiration)
	return nil
}

func (c *InMemoryCacheClient) Close() error {
	c.storage.Flush()
	return nil
}
