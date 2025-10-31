package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maisiq/go-auth-service/internal/service"
)

const UserEmailContextKey = "userEmail"

func AuthMiddleware(s service.SecretService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "empty auth header"})
			c.Abort()
			return
		}
		if !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "invalid token format"})
			c.Abort()
			return
		}

		t := strings.Split(token, " ")

		claims, err := s.ParseJWT(c, t[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": err.Error()})
			c.Abort()
			return
		}
		c.Set(UserEmailContextKey, claims.Email)
		c.Next()
	}
}
