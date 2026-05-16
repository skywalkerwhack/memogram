package app

import (
	"context"

	"github.com/usememos/memogram/internal/domain"
)

type TokenStore interface {
	GetUserAccessToken(userID int64) (string, bool)
	SetUserAccessToken(userID int64, accessToken string) error
	DeleteUserAccessToken(userID int64) (bool, error)
	CountUserAccessTokens() int
}

type Backend interface {
	BaseURL() string
	GetInstanceProfile(ctx context.Context) (*domain.InstanceProfile, error)
	GetCurrentUser(ctx context.Context, accessToken string) (*domain.User, error)
	CreateMemo(ctx context.Context, accessToken string, content string) (*domain.Memo, error)
	GetMemo(ctx context.Context, accessToken string, name string) (*domain.Memo, error)
	UpdateMemo(ctx context.Context, accessToken string, memo *domain.Memo) (*domain.Memo, error)
	DeleteMemo(ctx context.Context, accessToken string, name string) error
	SearchMemos(ctx context.Context, accessToken string, query string, creatorID *int64, limit int) ([]domain.Memo, error)
	UploadAttachment(ctx context.Context, accessToken string, memoName string, file domain.FilePayload) error
}
