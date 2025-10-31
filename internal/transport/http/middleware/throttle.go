package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func ThrottleMiddleware(limit, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(limit), burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{})
			c.Abort()
			return
		}
		c.Next()
	}
}
