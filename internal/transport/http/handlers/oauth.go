package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maisiq/go-auth-service/internal/configs"
	"github.com/maisiq/go-auth-service/internal/oauth"
	"github.com/maisiq/go-auth-service/internal/service"
)

// utils

func NewState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// handlers

type OAuthHandler struct {
	service        service.IOAuthService
	yandexProvider oauth.OAuthProvider
}

func NewOAuthHandler(s service.IOAuthService, yandex oauth.OAuthProvider) *OAuthHandler {
	return &OAuthHandler{
		service:        s,
		yandexProvider: yandex,
	}
}

func (h *OAuthHandler) Redirect(c *gin.Context) {
	providerName := c.Query("provider")
	state := NewState()

	var redirectURL string

	switch providerName {
	case string(oauth.YandexProvider):
		redirectURL = configs.GetConfig().Yandex.GetAuthorizeURL(state)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid provider in query string"})
		return
	}
	c.SetCookie("state", state, int((3 * time.Minute).Seconds()), "/", "localhost", false, true)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *OAuthHandler) YandexCallback(c *gin.Context) {
	// verify state
	qState := c.Query("state")
	cState, _ := c.Cookie("state")

	if qState != cState {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "bad request"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "bad request"})
		return
	}

	tokens, err := h.service.CreateTokens(c, h.yandexProvider, code)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.SetCookie(AccessTokenCookieKey, tokens.Access, int(service.AccessTokenTTL), "/", "localhost", false, true)
	c.SetCookie(RefreshTokenCookieKey, tokens.Refresh, int(service.RefreshTokenTTL), "/", "localhost", false, true)
	c.JSON(http.StatusCreated, tokens)
}
