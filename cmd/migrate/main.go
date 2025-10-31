package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/db/migrations"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func main() {
	cfgName, ok := os.LookupEnv("CONFIG_FILENAME")
	if !ok {
		panic("config path is not provided")
	}

	container := BuildContainer(fmt.Sprintf("./configs/%s", cfgName))
	log := wire.Get[*zap.SugaredLogger](container)
	db := wire.Get[*sqlx.DB](container)

	log.Info("run migrations")
	if err := makeMigrations(db.DB); err != nil {
		log.Errorw(
			"failed to run migrations",
			"error", err.Error(),
			"error_details", err,
		)
	}

	log.Info("successfullly migrated")
	container.ShutdownResources()
}

func makeMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		return err
	}
	return nil
}
