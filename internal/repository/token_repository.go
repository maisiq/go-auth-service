package repository

import (
	"context"
	"errors"
	"time"

	"github.com/maisiq/go-auth-service/internal/db"
	"github.com/redis/go-redis/v9"
)

type TokenRepository struct {
	client db.RedisClient
}

func NewTokenRepository(c db.RedisClient) *TokenRepository {
	return &TokenRepository{
		client: c,
	}
}

func (r *TokenRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	resp := r.client.Get(ctx, key)
	if err := resp.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}

	if err := resp.Scan(&value); err != nil {
		return "", err
	}
	return value, nil
}

func (r *TokenRepository) Add(ctx context.Context, key, value string, expiration time.Duration) error {
	resp := r.client.Set(ctx, key, value, expiration)
	if err := resp.Err(); err != nil {
		return err
	}
	return nil
}

func (r *TokenRepository) Delete(ctx context.Context, keys ...string) error {
	resp := r.client.Del(ctx, keys...)
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}

func (r *TokenRepository) Push(ctx context.Context, key string, values ...string) error {
	resp := r.client.SAdd(ctx, key, values)
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}

func (r *TokenRepository) List(ctx context.Context, key string) ([]string, error) {
	resp := r.client.SMembers(ctx, key)
	values, err := resp.Result()
	if err != nil {
		return nil, resp.Err()
	}
	return values, nil
}
