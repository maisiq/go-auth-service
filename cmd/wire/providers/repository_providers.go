package providers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/db"
	"github.com/maisiq/go-auth-service/internal/repository"
)

func UserRepoProvider(c *wire.DIContainer) repository.IUserRepository {
	db := wire.Get[*sqlx.DB](c)
	return repository.NewUserRepository(db)
}

func TokenRepoProvider(c *wire.DIContainer) repository.ITokenRepository {
	db := wire.Get[db.RedisClient](c)
	return repository.NewTokenRepository(db)
}

func SecretRepoProvider(c *wire.DIContainer) repository.SecretRepository {
	cfg := wire.Get[*configs.Config](c)
	client := wire.Get[*http.Client](c)
	return repository.NewVaultSecretRepository(cfg.Vault, client)
}
