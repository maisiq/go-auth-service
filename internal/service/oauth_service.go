package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/maisiq/go-auth-service/internal/domain"
	"github.com/maisiq/go-auth-service/internal/oauth"
	"github.com/maisiq/go-auth-service/internal/repository"
	"go.uber.org/zap"
)

type OAuthService struct {
	log        *zap.SugaredLogger
	tokenRepo  ITokenRepository
	userRepo   IUserRepository
	secretRepo repository.SecretRepository
}

func NewOAuthService(log *zap.SugaredLogger, userRepo IUserRepository, tokenRepo ITokenRepository, secretRepo repository.SecretRepository) *OAuthService {
	return &OAuthService{
		tokenRepo:  tokenRepo,
		userRepo:   userRepo,
		secretRepo: secretRepo,
		log:        log,
	}
}

func (s *OAuthService) CreateTokens(ctx context.Context, provider oauth.OAuthProvider, authorizationCode string) (*TokenPair, error) {
	data, err := provider.GetData(authorizationCode)
	if err != nil {
		return nil, err
	}
	u, err := s.getOrCreateUser(ctx, data.Email, data.UserID, data.Provider)
	if err != nil {
		return nil, err
	}
	access, err := createAccessToken(ctx, s.secretRepo, u.ID.String(), u.Email)
	if err != nil {
		s.log.Errorf("failed to create jwt token: %w", err)
		return nil, ErrInternal
	}
	refresh, err := createRefreshToken()
	if err != nil {
		s.log.Errorf("failed to create refresh token: %w", err)
		return nil, ErrInternal
	}
	tokens := TokenPair{
		Refresh: refresh,
		Access:  access,
	}

	if err := s.tokenRepo.Add(ctx, refresh, u.Email, RefreshTokenTTL); err != nil {
		s.log.Errorf("failed to add token to token repo: %w", err)
		return nil, ErrInternal
	}
	if err := s.tokenRepo.Push(ctx, u.Email, refresh); err != nil {
		s.log.Errorf("failed to add token to all user sessions: %w", err)
		return nil, ErrInternal
	}
	return &tokens, nil
}

func (s *OAuthService) getOrCreateUser(
	ctx context.Context,
	email, socialID string,
	socialProvider oauth.OAuthProviderT,
) (*domain.User, error) {
	normalizedEmail := normalizeEmail(email)

	u, err := s.userRepo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.log.Errorf("failed to get user: %w", err)
			return nil, ErrInternal
		}
		// if user doesn't exist
		id, _ := uuid.NewV7()
		newUser := domain.User{
			ID:             id,
			Email:          normalizedEmail,
			Role:           domain.UserRole,
			SocialAccount:  true,
			SocialID:       socialID,
			SocialProvider: string(socialProvider),
			CreatedAT:      time.Now(),
			LastLoogedAt:   time.Now(),
		}
		if err := s.userRepo.Add(ctx, newUser); err != nil {
			s.log.Errorw(
				"failed to add user to user repo",
				"function", "OAuthService.getOrCreateUser",
				"error", err.Error(),
				"error_details", err,
				"user", newUser,
			)
			return nil, ErrInternal
		}
		return &newUser, nil
	}
	// return if user exists
	return &u, nil
}
