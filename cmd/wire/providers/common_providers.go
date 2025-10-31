package providers

import (
	"context"

	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/logger"
	"go.uber.org/zap"
)

func ConfigProvider(c *wire.DIContainer) *configs.Config {
	path := wire.Get[wire.ConfigPath](c)

	ch := make(chan interface{})
	cfg, err := configs.LoadConfig(string(path), ch)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	c.RebuildOn(ctx, ch)
	return cfg
}

func LoggerProvider(c *wire.DIContainer) *zap.SugaredLogger {
	cfg := wire.Get[*configs.Config](c)
	return logger.InitLogger(cfg.App.Debug)
}
