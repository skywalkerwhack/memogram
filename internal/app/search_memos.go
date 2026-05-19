package app

import (
	"context"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (s *Service) SearchMemos(ctx context.Context, telegramUserID int64, query string, limit int) ([]domain.Memo, error) {
	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return nil, err
	}

	creatorID, err := s.lookupCreatorID(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return s.backend.SearchMemos(ctx, accessToken, query, creatorID, limit)
}
