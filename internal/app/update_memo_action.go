package app

import (
	"context"
	"fmt"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (s *Service) UpdateMemoAction(ctx context.Context, telegramUserID int64, action MemoAction, memoName string) (*domain.Memo, bool, error) {
	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return nil, false, err
	}

	if action == ActionDelete {
		if err := s.backend.DeleteMemo(ctx, accessToken, memoName); err != nil {
			return nil, false, err
		}
		return nil, true, nil
	}

	memo, err := s.backend.GetMemo(ctx, accessToken, memoName)
	if err != nil {
		return nil, false, err
	}

	switch action {
	case ActionPublic:
		memo.Visibility = domain.VisibilityPublic
	case ActionProtected:
		memo.Visibility = domain.VisibilityProtected
	case ActionPrivate:
		memo.Visibility = domain.VisibilityPrivate
	case ActionPin:
		memo.Pinned = !memo.Pinned
	default:
		return nil, false, fmt.Errorf("unknown action")
	}

	updatedMemo, err := s.backend.UpdateMemo(ctx, accessToken, memo)
	if err != nil {
		return nil, false, err
	}
	return updatedMemo, false, nil
}
