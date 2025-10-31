package router

import (
	"github.com/gin-gonic/gin"
	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/oauth"
	"github.com/maisiq/go-auth-service/internal/service"
	"github.com/maisiq/go-auth-service/internal/transport/http/handlers"
	"github.com/maisiq/go-auth-service/internal/transport/http/middleware"
	"go.opentelemetry.io/otel/trace"
)

type RouterParams struct {
	UserService    service.IUserService
	OAuthService   service.IOAuthService
	SecretService  service.SecretService
	YandexProvider oauth.OAuthProvider
	Config         *configs.AppConfig
	Tracer         trace.Tracer
}

func NewRouter(params *RouterParams) *gin.Engine {
	if params.Config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	uh := handlers.NewUserHadler(params.UserService)
	ah := handlers.NewOAuthHandler(params.OAuthService, params.YandexProvider)

	traced := r.Group("/")
	traced.Use(middleware.TracingMiddleware(params.Tracer))

	throttled := traced.Group("/")
	throttled.Use(middleware.ThrottleMiddleware(params.Config.Limiter.Limit, params.Config.Limiter.Burst))
	{
		throttled.POST("/create", uh.CreateUser)
		throttled.POST("/login", uh.AuthenticateUser)
		throttled.POST("/refresh", uh.Refresh)

		throttled.GET("/oauth/redirect", ah.Redirect)
		throttled.GET("/oauth/yandex/callback", ah.YandexCallback)
	}

	protected := throttled.Group("/")
	protected.Use(middleware.AuthMiddleware(params.SecretService))
	{
		protected.GET("/logs", uh.Logs)
		protected.POST("/logout", uh.Logout)
		protected.POST("/password", uh.UpdatePassword)
	}
	return r
}
