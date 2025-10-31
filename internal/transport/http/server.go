package xhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maisiq/go-auth-service/internal/configs"
	"go.uber.org/zap"
)

type ServerParams struct {
	Logger *zap.SugaredLogger
	Config *configs.Config
}

type Server struct {
	log    *zap.SugaredLogger
	config *configs.Config
	server *http.Server
}

func NewServer(params *ServerParams) *Server {
	return &Server{
		log:    params.Logger,
		config: params.Config,
	}
}

func (s *Server) GracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) RunWithHandler(handler http.Handler) {
	if s.config.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s.server = &http.Server{
		Addr:    s.config.App.Addr,
		Handler: handler,
	}
	s.log.Infow("starting http server", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.Errorf("server serving failed: %w", err)
	}
	s.log.Info("server is down")
}
