package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maisiq/go-auth-service/internal/cache"
	"github.com/maisiq/go-auth-service/internal/repository"
	"go.uber.org/zap"
)

type VaultService struct {
	repo  repository.SecretRepository
	log   *zap.SugaredLogger
	cache *cache.Cache
}

func NewVaultService(log *zap.SugaredLogger, repo repository.SecretRepository, cache *cache.Cache) *VaultService {
	return &VaultService{
		log:   log,
		repo:  repo,
		cache: cache,
	}
}

func (s *VaultService) ParseJWT(ctx context.Context, token string) (*AuthClaims, error) {
	claims := &AuthClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, ErrInvalidKID
		}

		keys, getErr := cache.GetOrSet(s.cache, ctx, JWTSingingKey, 5*time.Minute, func() (map[string]string, error) {
			return s.repo.GetPublicKeys(ctx, JWTSingingKey)
		})
		if getErr != nil {
			s.log.Errorf("failed to get public keys from repository: %w", getErr)
			return nil, ErrInternal
		}

		key, ok := keys[kid.(string)]
		if !ok {
			return nil, ErrInvalidKID
		}
		pk, err := parsePublicKey(key)
		if err != nil {
			s.log.Errorf("failed to parse pem public key: %w", err)
			return nil, ErrInternal
		}
		return pk, nil
	})

	if err != nil {
		return nil, err
	}

	if parsedToken.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken

}

func parsePublicKey(pemStr string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDSA public key: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return ecdsaPub, nil
}
