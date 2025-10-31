package providers

import (
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/cache"
)

func InMemoryCacheProvider(di *wire.DIContainer) *cache.Cache {
	client := cache.NewInMemoryCacheClient()
	return &cache.Cache{
		Client: client,
	}
}
