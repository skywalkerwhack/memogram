package app

import (
	"context"
	"fmt"
	"time"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (s *Service) CreateMemo(ctx context.Context, input CreateMemoInput) (*domain.Memo, error) {
	accessToken, err := s.requireAccessToken(input.UserID)
	if err != nil {
		return nil, err
	}

	content := input.Content
	if input.ForwardedFrom != nil {
		if input.ForwardedFrom.Username != "" {
			content = fmt.Sprintf("Forwarded from [%s](https://t.me/%s)\n%s", input.ForwardedFrom.Name, input.ForwardedFrom.Username, content)
		} else {
			content = fmt.Sprintf("Forwarded from %s\n%s", input.ForwardedFrom.Name, content)
		}
	}

	if input.AttachmentSet == "" {
		return s.backend.CreateMemo(ctx, accessToken, content)
	}

	s.mediaGroupMutex.Lock()
	defer s.mediaGroupMutex.Unlock()

	now := time.Now()
	s.deleteExpiredMediaGroups(now)

	if cached, ok := s.mediaGroupCache.Load(input.AttachmentSet); ok {
		entry := cached.(mediaGroupCacheEntry)
		if now.Before(entry.expiresAt) {
			return entry.memo, nil
		}
		s.mediaGroupCache.Delete(input.AttachmentSet)
	}

	memo, err := s.backend.CreateMemo(ctx, accessToken, content)
	if err != nil {
		return nil, err
	}
	s.mediaGroupCache.Store(input.AttachmentSet, mediaGroupCacheEntry{
		memo:      memo,
		expiresAt: now.Add(mediaGroupCacheTTL),
	})
	return memo, nil
}

func (s *Service) deleteExpiredMediaGroups(now time.Time) {
	s.mediaGroupCache.Range(func(key, value any) bool {
		entry, ok := value.(mediaGroupCacheEntry)
		if !ok || !now.Before(entry.expiresAt) {
			s.mediaGroupCache.Delete(key)
		}
		return true
	})
}

func (s *Service) AttachFile(ctx context.Context, telegramUserID int64, memoName string, file domain.FilePayload) error {
	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return err
	}
	return s.backend.UploadAttachment(ctx, accessToken, memoName, file)
}
