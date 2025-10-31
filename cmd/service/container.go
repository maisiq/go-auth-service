package main

import (
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/cmd/wire/providers"
)

func BuildContainer(configPath string) *wire.DIContainer {
	di := wire.New(configPath)

	wire.Provide(di, providers.ConfigProvider)
	wire.Provide(di, providers.LoggerProvider)

	// storages
	wire.Provide(di, providers.SQLDatabaseProvider)
	wire.Provide(di, providers.InMemoryDBProvider)
	wire.Provide(di, providers.HTTPClientProvider)

	// OAuth providers
	wire.Provide(di, providers.YandexOAuthProvider)

	// repositories
	wire.Provide(di, providers.UserRepoProvider)
	wire.Provide(di, providers.SecretRepoProvider)
	wire.Provide(di, providers.TokenRepoProvider)

	// cache
	wire.Provide(di, providers.InMemoryCacheProvider)

	// otel
	wire.Provide(di, providers.JaegerExporterProvider)
	wire.Provide(di, providers.TracerProvider)
	wire.ProvideNamed(di, "user-service", providers.UserServiceTracerProvider)
	wire.ProvideNamed(di, "http-server", providers.HTTPTracerProvider)

	// services
	wire.Provide(di, providers.OAuthServiceProvider)
	wire.Provide(di, providers.SecretServiceProvider)
	wire.Provide(di, providers.UserServiceProvider)

	// http server
	wire.Provide(di, providers.ServerParamsProvider)
	wire.Provide(di, providers.RouterParamsProvider)
	wire.Provide(di, providers.RouterProvider)
	wire.Provide(di, providers.HTTPServerProvider)
	return di
}
