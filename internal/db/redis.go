package db

import (
	"context"
	"time"

	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	Close() error
}

func NewRedisClient(cfg *configs.MemoryDBConfig) RedisClient {
	c := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return c
}
