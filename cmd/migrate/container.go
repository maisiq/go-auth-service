package main

import (
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/cmd/wire/providers"
)

func BuildContainer(configPath string) *wire.DIContainer {
	di := wire.New(configPath)
	wire.Provide(di, providers.ConfigProvider)
	wire.Provide(di, providers.LoggerProvider)

	wire.Provide(di, providers.SQLDatabaseProvider)
	return di
}
