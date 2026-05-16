package app

import (
	"time"

	"github.com/usememos/memogram/internal/domain"
)

type CreateMemoInput struct {
	UserID        int64
	Content       string
	AttachmentSet string
	ForwardedFrom *domain.ForwardInfo
}

type MemoAction string

const (
	ActionPublic    MemoAction = "public"
	ActionProtected MemoAction = "protected"
	ActionPrivate   MemoAction = "private"
	ActionPin       MemoAction = "pin"
	ActionDelete    MemoAction = "delete"
)

type StatusReport struct {
	ServerURL          string
	DataFile           string
	BackendLatency     time.Duration
	BackendAvailable   bool
	BackendError       string
	InstanceURL        string
	AllowedUsernames   int
	LinkedUsers        int
	AccountLinked      bool
	AccountTokenValid  bool
	AccountDisplayName string
}
