package service

import (
	"context"
	"time"

	"github.com/maisiq/go-auth-service/internal/domain"
	"github.com/maisiq/go-auth-service/internal/oauth"
)

type IOAuthService interface {
	CreateTokens(ctx context.Context, provider oauth.OAuthProvider, authorizationCode string) (*TokenPair, error)
}

type SecretService interface {
	ParseJWT(ctx context.Context, token string) (*AuthClaims, error)
}

type IUserService interface {
	CreateUser(ctx context.Context, email, password string) error
	Authenticate(ctx context.Context, email, password string) (*TokenPair, error)
	AddLog(ctx context.Context, userEmail, userAgent, IP string) error
	Logs(ctx context.Context, email string) ([]domain.UserLog, error)
	NewRefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string, fromAll bool) error
	UpdatePassword(ctx context.Context, email, old, new string) error
}

type IUserRepository interface {
	Add(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Logs(ctx context.Context, email string) ([]domain.UserLog, error)
	AddLog(ctx context.Context, log domain.UserLog) error
	Update(ctx context.Context, user domain.User) error
}

type ITokenRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, key, value string, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Push(ctx context.Context, key string, values ...string) error
	List(ctx context.Context, key string) ([]string, error)
}
