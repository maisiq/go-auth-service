package providers

import (
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/cache"
	"github.com/maisiq/go-auth-service/internal/repository"
	"github.com/maisiq/go-auth-service/internal/service"
	"github.com/maisiq/go-auth-service/pkg/resilience"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func UserServiceProvider(c *wire.DIContainer) service.IUserService {
	dbCB := resilience.NewCircuitBreakerDecorator("db", resilience.SqlDBClient)
	retryDB := resilience.NewRetryDecorator(resilience.RetryConfig{
		MaxAttempts: 3,
		Client:      resilience.SqlDBClient,
		MaxDelay:    2,
	})
	logger := wire.Get[*zap.SugaredLogger](c)
	userRepo := wire.Get[repository.IUserRepository](c)
	tokenRepo := wire.Get[repository.ITokenRepository](c)
	secretRepo := wire.Get[repository.SecretRepository](c)
	tracer := wire.GetNamed[trace.Tracer](c, "user-service")
	return service.NewUserService(logger, tracer, userRepo, tokenRepo, secretRepo, dbCB, retryDB)
}

func SecretServiceProvider(c *wire.DIContainer) service.SecretService {
	logger := wire.Get[*zap.SugaredLogger](c)
	secretRepo := wire.Get[repository.SecretRepository](c)
	cache := wire.Get[*cache.Cache](c)
	return service.NewVaultService(logger, secretRepo, cache)
}

func OAuthServiceProvider(c *wire.DIContainer) service.IOAuthService {
	logger := wire.Get[*zap.SugaredLogger](c)
	userRepo := wire.Get[repository.IUserRepository](c)
	tokenRepo := wire.Get[repository.ITokenRepository](c)
	secretRepo := wire.Get[repository.SecretRepository](c)
	return service.NewOAuthService(logger, userRepo, tokenRepo, secretRepo)
}
