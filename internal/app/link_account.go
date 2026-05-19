package app

import (
	"context"
	"fmt"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (s *Service) LinkAccount(ctx context.Context, telegramUserID int64, accessToken string) (string, error) {
	user, err := s.backend.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
	}
	if err := s.store.SetUserAccessToken(telegramUserID, accessToken); err != nil {
		return "", fmt.Errorf("save access token: %w", err)
	}

	return displayNameOf(user), nil
}

func (s *Service) UnlinkAccount(telegramUserID int64) (bool, error) {
	return s.store.DeleteUserAccessToken(telegramUserID)
}

func displayNameOf(user *domain.User) string {
	if user == nil {
		return ""
	}

	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.Username
	}
	if displayName == "" {
		displayName = user.Name
	}
	return displayName
}
