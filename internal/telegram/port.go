package telegram

import (
	"context"

	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

type Service interface {
	Start(ctx context.Context)
	IsUserAllowed(username string) bool
	IsUserAdmin(username string) bool
	LinkAccount(ctx context.Context, userID int64, accessToken string) (string, error)
	UnlinkAccount(userID int64) (bool, error)
	CreateMemo(ctx context.Context, input app.CreateMemoInput) (*domain.Memo, error)
	AttachFile(ctx context.Context, userID int64, memoName string, file domain.FilePayload) error
	SearchMemos(ctx context.Context, userID int64, query string, limit int) ([]domain.Memo, error)
	GetStatus(ctx context.Context, userID int64) app.StatusReport
	GetHealth(ctx context.Context) app.HealthReport
	UpdateMemoAction(ctx context.Context, userID int64, action app.MemoAction, memoName string) (*domain.Memo, bool, error)
	MemoBaseURL() string
}
