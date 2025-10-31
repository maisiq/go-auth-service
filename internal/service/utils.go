package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/maisiq/go-auth-service/internal/repository"
	"go.opentelemetry.io/otel"
)

const AccessTokenTTL = 5 * time.Minute
const RefreshTokenTTL = 24 * time.Hour

var JWTSingingKey = "jwt-key"

type AuthClaims struct {
	Role  string `json:"role"`
	Email string `json:"email"`

	jwt.RegisteredClaims
}

func hashPassword(plain string) (string, error) {
	return argon2id.CreateHash(plain, argon2id.DefaultParams)
}

func authenticate(plain string, hashedPassword string) (bool, error) {
	return argon2id.ComparePasswordAndHash(plain, hashedPassword)
}

func createAccessToken(ctx context.Context, repo repository.SecretRepository, userID, email string) (string, error) {
	tr := otel.GetTracerProvider().Tracer("gin-server")
	ctx, span := tr.Start(ctx, "createAccessToken")
	defer span.End()

	claims := AuthClaims{
		"user",
		email,
		jwt.RegisteredClaims{
			Issuer:    "auth-service",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	kid, err := repo.GetKID(ctx, JWTSingingKey)
	if err != nil {
		return "", fmt.Errorf("failed to get kid from secret repository: %w", err)
	}
	token.Header["kid"] = kid

	ss, err := token.SigningString()
	if err != nil {
		return "", fmt.Errorf("failed to get signing string")
	}

	signedString, err := repo.SignJWT(ctx, ss, JWTSingingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign data via secret repository: %w", err)
	}
	return signedString, nil
}

func createRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(email)
}
