package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maisiq/go-auth-service/internal/logger"
	"github.com/maisiq/go-auth-service/internal/service"
	"github.com/maisiq/go-auth-service/internal/transport/http/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type AutheticateUserRequest struct {
	CreateUserRequest
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gte=6"`
}

type ResponseWithError struct {
	Detail string `json:"detail"`
}

type UserHadlerGin struct {
	service service.IUserService
}

func NewUserHadler(s service.IUserService) *UserHadlerGin {
	return &UserHadlerGin{
		service: s,
	}
}

func (h *UserHadlerGin) CreateUser(c *gin.Context) {
	var s CreateUserRequest

	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithError{Detail: "invalid body"})
		return
	}

	err := h.service.CreateUser(c, s.Email, s.Password)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "user with this email already exists"})
			return
		}
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func (h *UserHadlerGin) AuthenticateUser(c *gin.Context) {
	var s AutheticateUserRequest

	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("handler", "authenticate_user"))
	defer span.End()

	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithError{Detail: "invalid body"})
		return
	}

	token, err := h.service.Authenticate(c.Request.Context(), s.Email, s.Password)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": "user not found"})
			return
		} else if errors.Is(err, service.ErrBadCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "email-password pair don't match"})
			return
		}
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	err = h.service.AddLog(c, s.Email, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		logger.GetLogger().Error(err)
	}

	c.JSON(http.StatusOK, token)
}

func (h *UserHadlerGin) Logs(c *gin.Context) {
	email := c.GetString(middleware.UserEmailContextKey)
	if email == "" {
		c.JSON(http.StatusInternalServerError, gin.H{})
	}

	logs, err := h.service.Logs(c, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *UserHadlerGin) Refresh(c *gin.Context) {
	t, err := c.Cookie(RefreshTokenCookieKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "authorization cookie was not provided"})
		return
	}

	tokens, err := h.service.NewRefreshToken(c, t)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		} else if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "internal error"})
		}
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *UserHadlerGin) Logout(c *gin.Context) {
	t, err := c.Cookie(RefreshTokenCookieKey)
	if err != nil || t == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "got no refresh token in cookie"})
		return
	}

	var all bool
	fromAll := c.Query("all")
	if fromAll != "" {
		all = true
	}

	if err := h.service.Logout(c, t, all); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

type UpdatePasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OldPassword string `json:"old_password" binding:"required"`
	Password    string `json:"password" binding:"required,nefield=OldPassword"`
}

func (h *UserHadlerGin) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithError{Detail: err.Error()})
		return
	}

	if err := h.service.UpdatePassword(c, req.Email, req.OldPassword, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithError{Detail: err.Error()})
		return
	}
	c.SetCookie(RefreshTokenCookieKey, "", 0, "/", "localhost", false, true)
	c.SetCookie(AccessTokenCookieKey, "", 0, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{})
}
