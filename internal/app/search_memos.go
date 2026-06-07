package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (s *Service) SearchMemos(ctx context.Context, telegramUserID int64, query string, limit int) ([]domain.Memo, error) {
	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return nil, err
	}

	user, err := s.backend.GetCurrentUser(ctx, accessToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return nil, fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
		}
		return nil, err
	}

	var creatorID *int64
	if user != nil {
		if tokens, err := domain.GetNameParentTokens(user.Name, "users/"); err == nil && len(tokens) == 1 {
			if parsedUserID, err := strconv.ParseInt(tokens[0], 10, 64); err == nil {
				creatorID = &parsedUserID
			}
		}
	}

	return s.backend.SearchMemos(ctx, accessToken, query, creatorID, limit)
}
