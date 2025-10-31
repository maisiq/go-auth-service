package providers

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/db"
)

type Constructor[T any] func(*wire.DIContainer) T

func SQLDatabaseProvider(c *wire.DIContainer) *sqlx.DB {
	cfg := wire.Get[*configs.Config](c)
	db, err := db.NewSQLDatabase(cfg)
	if err != nil {
		panic(err)
	}
	c.AddToCloser(func() error {
		return db.Close()
	})

	return db
}

func InMemoryDBProvider(c *wire.DIContainer) db.RedisClient {
	cfg := wire.Get[*configs.Config](c)
	client := db.NewRedisClient(cfg.MemoryDB)

	c.AddToCloser(func() error {
		return client.Close()
	})

	return client
}

func HTTPClientProvider(c *wire.DIContainer) *http.Client {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	return client
}
