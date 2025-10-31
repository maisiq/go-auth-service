package cache

import (
	"context"
	"encoding/json"
	"time"
)

type CacheClient interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Close() error
}

type Cache struct {
	Client CacheClient
}

func GetOrSet[T any](c *Cache, ctx context.Context, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	var empty T

	val, clientErr := c.Client.Get(ctx, key)

	if clientErr == nil {
		var dto T
		if convertErr := json.Unmarshal(val, &dto); convertErr == nil {
			return dto, nil
		} else {
			return empty, convertErr
		}
	}

	if clientErr != ErrNotFound {
		return empty, clientErr
	}

	dto, storageErr := fetch()
	if storageErr != nil {
		return empty, storageErr
	}

	data, err := json.Marshal(dto)
	if err == nil {
		_ = c.Client.Set(ctx, key, data, ttl)
	}

	return dto, nil
}
