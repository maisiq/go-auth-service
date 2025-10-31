package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/maisiq/go-auth-service/internal/domain"
	"github.com/maisiq/go-auth-service/internal/repository"
	"github.com/maisiq/go-auth-service/pkg/resilience"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type TokenPair struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

type UserService struct {
	log        *zap.SugaredLogger
	tracer     trace.Tracer
	userRepo   IUserRepository
	tokenRepo  ITokenRepository
	secretRepo repository.SecretRepository

	dbCB  *gobreaker.CircuitBreaker
	retry *resilience.Retry
}

func NewUserService(
	log *zap.SugaredLogger,
	tracer trace.Tracer,
	userRepo IUserRepository,
	tokenRepo ITokenRepository,
	secretRepo repository.SecretRepository,
	dbCB *gobreaker.CircuitBreaker,
	retry *resilience.Retry,
) *UserService {
	return &UserService{
		log:        log,
		tracer:     tracer,
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		secretRepo: secretRepo,
		dbCB:       dbCB,
		retry:      retry,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email, password string) error {
	dctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var err error

	hashedPwd, err := hashPassword(password)
	if err != nil {
		s.log.Errorf("failed to hash password: %w", err)
		return ErrInternal
	}

	nemail := normalizeEmail(email)

	id, _ := uuid.NewV7()
	u := domain.User{
		ID:             id,
		Email:          nemail,
		Role:           domain.UserRole,
		HashedPassword: hashedPwd,
		CreatedAT:      time.Now(),
	}

	_, err = s.dbCB.Execute(func() (interface{}, error) {
		s.log.Debug("CB called")
		err = s.retry.Call(func() error {
			s.log.Debug("retry called")
			return s.userRepo.Add(dctx, u)
		})
		s.log.Debugf("retry returns err: %v", err)
		return nil, err
	})
	s.log.Debugf("CB returns err: %v", err)

	if errors.Is(err, repository.ErrAlreadyExists) {
		return ErrAlreadyExists
	}
	if err != nil {
		s.log.Errorf("failed to add user to db: %w\n", err)
		return ErrInternal
	}

	return nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*TokenPair, error) {
	sctx, span := s.tracer.Start(ctx, "UserService.Authenticate")
	defer span.End()

	var traceID string
	if span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
	}
	span.SetAttributes(attribute.String("user.email", email))

	span.AddEvent("get user from repo")
	user, err := s.userRepo.GetByEmail(sctx, email)

	if err != nil {
		span.RecordError(err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		span.SetStatus(codes.Error, "failed to retrieve user")
		s.log.Errorw("failed to retrieve user",
			"user_email", email,
			"error", err,
			"trace_id", traceID,
		)
		return nil, ErrInternal
	}

	span.AddEvent("compare password and hash")
	ok, err := authenticate(password, user.HashedPassword)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to compare plain password with hash")
		s.log.Errorw("failed to authenticate user with email",
			"user_email", email,
			"error", err,
			"trace_id", traceID,
		)
		return nil, ErrInternal
	}

	if !ok {
		return nil, ErrBadCredentials
	}

	span.AddEvent("create access token")
	token, err := createAccessToken(sctx, s.secretRepo, user.ID.String(), user.Email)
	if err != nil {
		errMsg := "failed to create access token"
		span.RecordError(err)
		span.SetStatus(codes.Error, errMsg)
		s.log.Errorw(errMsg,
			"error", err,
			"trace_id", traceID,
		)
		return nil, err
	}

	refresh, err := createRefreshToken()
	if err != nil {
		errMsg := "failed to create refresh token"
		span.RecordError(err)
		span.SetStatus(codes.Error, errMsg)
		s.log.Errorw(errMsg,
			"error", err,
			"trace_id", traceID,
		)
		return nil, err
	}

	span.AddEvent("save refresh token")
	if err = s.tokenRepo.Add(sctx, refresh, user.Email, RefreshTokenTTL); err != nil {
		errMsg := "failed to save refresh token"
		span.RecordError(err)
		span.SetStatus(codes.Error, errMsg)
		s.log.Errorf(errMsg,
			"error", err,
			"trace_id", traceID,
		)
		return nil, err
	}

	if err = s.tokenRepo.Push(sctx, user.Email, refresh); err != nil {
		errMsg := "failed to push refresh token to all user's token"
		span.RecordError(err)
		span.SetStatus(codes.Error, errMsg)
		s.log.Errorf(errMsg,
			"error", err,
			"trace_id", traceID,
		)
		return nil, err
	}

	return &TokenPair{
		Access:  token,
		Refresh: refresh,
	}, nil
}

func (s *UserService) AddLog(ctx context.Context, userEmail, userAgent, IP string) error {
	dctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	id, _ := uuid.NewV7()
	log := domain.UserLog{
		ID:        id,
		UserEmail: userEmail,
		UserAgent: userAgent,
		IP:        IP,
		LoggedAt:  time.Now(),
	}
	if err := s.userRepo.AddLog(dctx, log); err != nil {
		s.log.Errorf("failed to add user log: %w", err)
		return ErrInternal
	}
	return nil
}

func (s *UserService) Logs(ctx context.Context, email string) ([]domain.UserLog, error) {
	dctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	normalizedEmail := normalizeEmail(email)
	logs, err := s.userRepo.Logs(dctx, normalizedEmail)
	if err != nil {
		s.log.Errorf("failed to get user log: %w", err)
		return nil, err
	}
	return logs, nil
}

func (s *UserService) NewRefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	email, err := s.tokenRepo.Get(ctx, refreshToken)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		s.log.Errorf("failed to get refresh token: %w", err)
		return nil, ErrInternal
	}

	user, err := s.userRepo.GetByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.log.Errorf("failed to get user: %w", err)
		return nil, ErrInternal
	}

	newAccess, err := createAccessToken(ctx, s.secretRepo, user.Email, user.Email)
	if err != nil {
		s.log.Errorf("failed to generate jwt token: %w", err)
		return nil, ErrInternal
	}

	newRefresh, err := createRefreshToken()
	if err != nil {
		s.log.Errorf("failed to generate refresh token: %w", err)
		return nil, ErrInternal
	}

	if err := s.tokenRepo.Delete(ctx, refreshToken); err != nil {
		s.log.Errorf("failed to delete old refresh token: %w", err)
		return nil, ErrInternal
	}

	s.tokenRepo.Add(ctx, newRefresh, email, RefreshTokenTTL)

	return &TokenPair{
		Access:  newAccess,
		Refresh: newRefresh,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, refreshToken string, fromAll bool) error {
	var tokens = []string{refreshToken}

	if fromAll {
		email, err := s.tokenRepo.Get(ctx, refreshToken)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrNotFound
			}
			s.log.Errorf("failed to delete tokens: %w", err)
			return ErrInternal
		}

		t, err := s.tokenRepo.List(ctx, email)
		if err != nil {
			return err
		}
		tokens = append(tokens, email)
		tokens = append(tokens, t...)
	}
	if err := s.tokenRepo.Delete(ctx, tokens...); err != nil {
		s.log.Errorf("failed to delete tokens: %w", err)
		return ErrInternal
	}
	return nil
}

func (s *UserService) UpdatePassword(ctx context.Context, email, old, new string) error {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		s.log.Errorf("failed to get user: %w", err)
		return ErrInternal
	}
	ok, err := authenticate(old, u.HashedPassword)
	if err != nil {
		s.log.Errorf("failed to autheticate user: %w", err)
		return ErrInternal
	}
	if !ok {
		return ErrBadCredentials
	}

	newHashedPassword, err := hashPassword(new)
	if err != nil {
		s.log.Errorf("failed to hash password: %w", err)
		return ErrInternal
	}
	u.HashedPassword = newHashedPassword
	if err := s.userRepo.Update(ctx, u); err != nil {
		s.log.Errorf("failed to update user: %w", err)
		return ErrInternal
	}

	//TODO: logout from account(s), but it needs refresh token
	return nil
}
