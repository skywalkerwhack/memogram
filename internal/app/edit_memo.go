package app

import (
	"context"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (s *Service) BeginMemoEdit(ctx context.Context, telegramUserID int64, memoName string) (*domain.Memo, error) {
	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return nil, err
	}

	memo, err := s.backend.GetMemo(ctx, accessToken, memoName)
	if err != nil {
		return nil, err
	}

	s.pendingEdits.Store(telegramUserID, memoName)
	return memo, nil
}

func (s *Service) HasPendingMemoEdit(telegramUserID int64) bool {
	_, ok := s.pendingEdits.Load(telegramUserID)
	return ok
}

func (s *Service) CancelMemoEdit(telegramUserID int64) bool {
	_, ok := s.pendingEdits.Load(telegramUserID)
	if ok {
		s.pendingEdits.Delete(telegramUserID)
	}
	return ok
}

func (s *Service) UpdatePendingMemoContent(ctx context.Context, telegramUserID int64, content string) (*domain.Memo, error) {
	memoNameValue, ok := s.pendingEdits.Load(telegramUserID)
	if !ok {
		return nil, domain.ErrMemoEditNotFound
	}

	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return nil, err
	}

	memoName := memoNameValue.(string)
	memo, err := s.backend.GetMemo(ctx, accessToken, memoName)
	if err != nil {
		return nil, err
	}

	memo.Content = content
	updatedMemo, err := s.backend.UpdateMemo(ctx, accessToken, memo)
	if err != nil {
		return nil, err
	}

	s.pendingEdits.Delete(telegramUserID)
	return updatedMemo, nil
}
