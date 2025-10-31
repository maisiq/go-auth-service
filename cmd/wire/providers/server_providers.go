package providers

import (
	"github.com/gin-gonic/gin"
	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/oauth"
	"github.com/maisiq/go-auth-service/internal/service"
	xhttp "github.com/maisiq/go-auth-service/internal/transport/http"
	"github.com/maisiq/go-auth-service/internal/transport/http/router"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func ServerParamsProvider(di *wire.DIContainer) *xhttp.ServerParams {
	params := &xhttp.ServerParams{
		Logger: wire.Get[*zap.SugaredLogger](di),
		Config: wire.Get[*configs.Config](di),
	}
	return params
}

func RouterParamsProvider(di *wire.DIContainer) *router.RouterParams {
	cfg := wire.Get[*configs.Config](di)
	params := &router.RouterParams{
		UserService:    wire.Get[service.IUserService](di),
		OAuthService:   wire.Get[service.IOAuthService](di),
		SecretService:  wire.Get[service.SecretService](di),
		YandexProvider: wire.Get[oauth.OAuthProvider](di),
		Config:         cfg.App,
		Tracer:         wire.GetNamed[trace.Tracer](di, "http-server"),
	}
	return params
}

func RouterProvider(di *wire.DIContainer) *gin.Engine {
	params := wire.Get[*router.RouterParams](di)
	return router.NewRouter(params)
}

func HTTPServerProvider(di *wire.DIContainer) *xhttp.Server {
	params := wire.Get[*xhttp.ServerParams](di)

	srv := xhttp.NewServer(params)
	di.AddToCloser(func() error {
		return srv.GracefulShutdown()
	})
	return srv
}
