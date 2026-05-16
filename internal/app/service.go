package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/usememos/memogram/internal/domain"
)

type Service struct {
	backend Backend
	store   TokenStore

	serverURL        string
	dataFile         string
	allowedUsernames map[string]struct{}

	mediaGroupCache sync.Map
	mediaGroupMutex sync.Mutex

	instanceProfile *domain.InstanceProfile
}

func NewService(backend Backend, store TokenStore, dataFile string, allowedUsernames []string) *Service {
	allowedSet := make(map[string]struct{}, len(allowedUsernames))
	for _, username := range allowedUsernames {
		allowedSet[strings.ToLower(strings.TrimSpace(username))] = struct{}{}
	}

	return &Service{
		backend:          backend,
		store:            store,
		serverURL:        backend.BaseURL(),
		dataFile:         dataFile,
		allowedUsernames: allowedSet,
	}
}

func (s *Service) Start(ctx context.Context) {
	profile, err := s.backend.GetInstanceProfile(ctx)
	if err == nil {
		s.instanceProfile = profile
	}
}

func (s *Service) IsUserAllowed(username string) bool {
	if len(s.allowedUsernames) == 0 {
		return true
	}
	if username == "" {
		return false
	}
	_, ok := s.allowedUsernames[strings.ToLower(strings.TrimSpace(username))]
	return ok
}

func (s *Service) LinkAccount(ctx context.Context, telegramUserID int64, accessToken string) (string, error) {
	user, err := s.backend.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
	}
	if err := s.store.SetUserAccessToken(telegramUserID, accessToken); err != nil {
		return "", fmt.Errorf("save access token: %w", err)
	}

	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.Username
	}
	if displayName == "" {
		displayName = user.Name
	}
	return displayName, nil
}

func (s *Service) UnlinkAccount(telegramUserID int64) (bool, error) {
	return s.store.DeleteUserAccessToken(telegramUserID)
}

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

	if cached, ok := s.mediaGroupCache.Load(input.AttachmentSet); ok {
		return cached.(*domain.Memo), nil
	}

	memo, err := s.backend.CreateMemo(ctx, accessToken, content)
	if err != nil {
		return nil, err
	}
	s.mediaGroupCache.Store(input.AttachmentSet, memo)
	return memo, nil
}

func (s *Service) AttachFile(ctx context.Context, telegramUserID int64, memoName string, file domain.FilePayload) error {
	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return err
	}
	return s.backend.UploadAttachment(ctx, accessToken, memoName, file)
}

func (s *Service) SearchMemos(ctx context.Context, telegramUserID int64, query string, limit int) ([]domain.Memo, error) {
	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return nil, err
	}

	user, err := s.backend.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
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

func (s *Service) GetStatus(ctx context.Context, telegramUserID int64) StatusReport {
	backendStatus := ProbeBackendLatency(ctx, s.backend)
	report := StatusReport{
		ServerURL:        s.serverURL,
		DataFile:         s.dataFile,
		BackendLatency:   backendStatus.Latency,
		BackendAvailable: backendStatus.Err == nil,
		LinkedUsers:      s.store.CountUserAccessTokens(),
	}
	if backendStatus.Err != nil {
		report.BackendError = sanitizeBackendError(backendStatus.Err)
	}

	if s.instanceProfile != nil {
		report.InstanceURL = s.instanceProfile.InstanceURL
	}

	if len(s.allowedUsernames) > 0 {
		report.AllowedUsernames = len(s.allowedUsernames)
	}

	accessToken, ok := s.store.GetUserAccessToken(telegramUserID)
	if !ok {
		return report
	}
	report.AccountLinked = true

	user, err := s.backend.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return report
	}
	report.AccountTokenValid = true
	report.AccountDisplayName = displayNameOf(user)
	return report
}

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

func (s *Service) MemoBaseURL() string {
	if s.instanceProfile != nil && s.instanceProfile.InstanceURL != "" {
		return s.instanceProfile.InstanceURL
	}
	return s.serverURL
}

func (s *Service) requireAccessToken(telegramUserID int64) (string, error) {
	accessToken, ok := s.store.GetUserAccessToken(telegramUserID)
	if !ok {
		return "", domain.ErrAccountNotLinked
	}
	return accessToken, nil
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
