package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/maisiq/go-auth-service/internal/domain"
	"github.com/maisiq/go-auth-service/internal/repository/mocks"
	"github.com/maisiq/go-auth-service/internal/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetLogs(t *testing.T) {
	logger := zap.NewExample()
	userRepo := mocks.NewIUserRepositoryMock(t)
	tokenRepo := mocks.NewITokenRepositoryMock(t)
	secretRepo := mocks.NewSecretRepositoryMock(t)

	ctx := context.Background()

	userService := service.NewUserService(logger.Sugar(), nil, userRepo, tokenRepo, secretRepo, nil, nil)

	t.Run("logs returns logs", func(t *testing.T) {
		email := "exAmplE@gmail.com"

		userRepo.LogsMock.Expect(minimock.AnyContext, strings.ToLower(email)).Return(
			[]domain.UserLog{}, nil,
		)
		result, err := userService.Logs(ctx, email)

		require.ErrorIs(t, err, nil)
		require.Equal(t, 0, len(result))
		require.Equal(t, make([]domain.UserLog, 0), result)
	})
}
