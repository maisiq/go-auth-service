package providers

import (
	"net/http"

	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/oauth"
	"github.com/maisiq/go-auth-service/internal/oauth/providers/yandex"
)

func YandexOAuthProvider(c *wire.DIContainer) oauth.OAuthProvider {
	cfg := wire.Get[*configs.Config](c)
	client := wire.Get[*http.Client](c)
	return yandex.NewYandexOAuthProvider(cfg.Yandex, client)
}
