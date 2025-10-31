package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(tr trace.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tr.Start(c.Request.Context(), c.FullPath())
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("server.address", c.Request.Host),
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		span.SetAttributes(
			attribute.Int("http.status", c.Writer.Status()),
		)

	}
}
