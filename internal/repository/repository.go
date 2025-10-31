package repository

import (
	"context"
	"time"

	"github.com/maisiq/go-auth-service/internal/domain"
)

//go:generate minimock -i IUserRepository,SecretRepository,ITokenRepository -o ./mocks/ -s "_mock.go"
type IUserRepository interface {
	Add(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Logs(ctx context.Context, email string) ([]domain.UserLog, error)
	AddLog(ctx context.Context, log domain.UserLog) error
	Update(ctx context.Context, user domain.User) error
}

type SecretRepository interface {
	GetKID(ctx context.Context, keyName string) (string, error)
	GetPublicKeys(ctx context.Context, keyName string) (map[string]string, error)
	SignJWT(ctx context.Context, data string, keyName string) (string, error)
}

type ITokenRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, key, value string, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Push(ctx context.Context, key string, values ...string) error
	List(ctx context.Context, key string) ([]string, error)
}
