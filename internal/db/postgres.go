package db

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/maisiq/go-auth-service/internal/configs"
)

func NewSQLDatabase(c *configs.Config) (*sqlx.DB, error) {
	dsn := c.Database.DSN
	if dsn == "" {
		panic("DSN is not provided")
	}

	client, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	client.SetMaxOpenConns(c.Database.MaxConn)

	if err = client.Ping(); err != nil {
		return nil, err
	}
	return client, nil
}
