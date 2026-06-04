package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

type fakeBackend struct {
	baseURL string

	getInstanceProfile func(context.Context) (*domain.InstanceProfile, error)
	getCurrentUser     func(context.Context, string) (*domain.User, error)
	createMemo         func(context.Context, string, string) (*domain.Memo, error)
	getMemo            func(context.Context, string, string) (*domain.Memo, error)
	updateMemo         func(context.Context, string, *domain.Memo) (*domain.Memo, error)
	deleteMemo         func(context.Context, string, string) error
	searchMemos        func(context.Context, string, string, *int64, int) ([]domain.Memo, error)
	uploadAttachment   func(context.Context, string, string, domain.FilePayload) error
}

func (b fakeBackend) BaseURL() string { return b.baseURL }

func (b fakeBackend) GetInstanceProfile(ctx context.Context) (*domain.InstanceProfile, error) {
	if b.getInstanceProfile != nil {
		return b.getInstanceProfile(ctx)
	}
	return nil, errors.New("not implemented")
}

func (b fakeBackend) GetCurrentUser(ctx context.Context, accessToken string) (*domain.User, error) {
	if b.getCurrentUser != nil {
		return b.getCurrentUser(ctx, accessToken)
	}
	return nil, errors.New("not implemented")
}

func (b fakeBackend) CreateMemo(ctx context.Context, accessToken string, content string) (*domain.Memo, error) {
	if b.createMemo != nil {
		return b.createMemo(ctx, accessToken, content)
	}
	return nil, errors.New("not implemented")
}

func (b fakeBackend) GetMemo(ctx context.Context, accessToken string, name string) (*domain.Memo, error) {
	if b.getMemo != nil {
		return b.getMemo(ctx, accessToken, name)
	}
	return nil, errors.New("not implemented")
}

func (b fakeBackend) UpdateMemo(ctx context.Context, accessToken string, memo *domain.Memo) (*domain.Memo, error) {
	if b.updateMemo != nil {
		return b.updateMemo(ctx, accessToken, memo)
	}
	return nil, errors.New("not implemented")
}

func (b fakeBackend) DeleteMemo(ctx context.Context, accessToken string, name string) error {
	if b.deleteMemo != nil {
		return b.deleteMemo(ctx, accessToken, name)
	}
	return errors.New("not implemented")
}

func (b fakeBackend) SearchMemos(ctx context.Context, accessToken string, query string, creatorID *int64, limit int) ([]domain.Memo, error) {
	if b.searchMemos != nil {
		return b.searchMemos(ctx, accessToken, query, creatorID, limit)
	}
	return nil, errors.New("not implemented")
}

func (b fakeBackend) UploadAttachment(ctx context.Context, accessToken string, memoName string, file domain.FilePayload) error {
	if b.uploadAttachment != nil {
		return b.uploadAttachment(ctx, accessToken, memoName, file)
	}
	return errors.New("not implemented")
}

type fakeTokenStore struct {
	tokens    map[int64]string
	setErr    error
	deleteErr error
}

func (s *fakeTokenStore) GetUserAccessToken(userID int64) (string, bool) {
	token, ok := s.tokens[userID]
	return token, ok
}

func (s *fakeTokenStore) SetUserAccessToken(userID int64, accessToken string) error {
	if s.setErr != nil {
		return s.setErr
	}
	if s.tokens == nil {
		s.tokens = map[int64]string{}
	}
	s.tokens[userID] = accessToken
	return nil
}

func (s *fakeTokenStore) DeleteUserAccessToken(userID int64) (bool, error) {
	if s.deleteErr != nil {
		return false, s.deleteErr
	}
	if _, ok := s.tokens[userID]; !ok {
		return false, nil
	}
	delete(s.tokens, userID)
	return true, nil
}

func (s *fakeTokenStore) CountUserAccessTokens() int {
	return len(s.tokens)
}

func TestServiceIsUserAllowed(t *testing.T) {
	service := NewService(fakeBackend{baseURL: "https://example.test"}, &fakeTokenStore{}, "data.txt", []string{"Alice", " Bob "}, nil)

	if !service.IsUserAllowed("alice") {
		t.Fatal("expected alice to be allowed")
	}
	if !service.IsUserAllowed("BOB") {
		t.Fatal("expected bob to be allowed")
	}
	if service.IsUserAllowed("") {
		t.Fatal("expected empty username to be rejected when allowlist is configured")
	}
	if service.IsUserAllowed("mallory") {
		t.Fatal("expected mallory to be rejected")
	}
}

func TestServiceLinkAccountStoresTokenAndUsesDisplayFallbacks(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{}}
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getCurrentUser: func(context.Context, string) (*domain.User, error) {
			return &domain.User{Username: "fallback-user"}, nil
		},
	}, store, "data.txt", nil, nil)

	displayName, err := service.LinkAccount(context.Background(), 42, "secret")
	if err != nil {
		t.Fatalf("LinkAccount returned error: %v", err)
	}
	if displayName != "fallback-user" {
		t.Fatalf("expected username fallback, got %q", displayName)
	}
	if got := store.tokens[42]; got != "secret" {
		t.Fatalf("expected token to be stored, got %q", got)
	}
}

func TestServiceLinkAccountInvalidToken(t *testing.T) {
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getCurrentUser: func(context.Context, string) (*domain.User, error) {
			return nil, errors.New("unauthorized")
		},
	}, &fakeTokenStore{}, "data.txt", nil, nil)

	_, err := service.LinkAccount(context.Background(), 42, "bad-token")
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceCreateMemoWithForwardedContent(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	var gotContent string
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		createMemo: func(_ context.Context, accessToken string, content string) (*domain.Memo, error) {
			if accessToken != "token" {
				t.Fatalf("expected access token to be used, got %q", accessToken)
			}
			gotContent = content
			return &domain.Memo{Name: "memos/1", Content: content}, nil
		},
	}, store, "data.txt", nil, nil)

	memo, err := service.CreateMemo(context.Background(), CreateMemoInput{
		UserID:  7,
		Content: "hello",
		ForwardedFrom: &domain.ForwardInfo{
			Name:     "Alice",
			Username: "alice",
		},
	})
	if err != nil {
		t.Fatalf("CreateMemo returned error: %v", err)
	}

	want := "Forwarded from [Alice](https://t.me/alice)\nhello"
	if gotContent != want {
		t.Fatalf("unexpected content:\nwant: %q\ngot:  %q", want, gotContent)
	}
	if memo.Content != want {
		t.Fatalf("expected memo content to match, got %q", memo.Content)
	}
}

func TestServiceCreateMemoUsesMediaGroupCache(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	createCalls := 0
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		createMemo: func(context.Context, string, string) (*domain.Memo, error) {
			createCalls++
			return &domain.Memo{Name: fmt.Sprintf("memos/%d", createCalls)}, nil
		},
	}, store, "data.txt", nil, nil)

	first, err := service.CreateMemo(context.Background(), CreateMemoInput{
		UserID:        7,
		Content:       "one",
		AttachmentSet: "album-1",
	})
	if err != nil {
		t.Fatalf("first CreateMemo returned error: %v", err)
	}
	second, err := service.CreateMemo(context.Background(), CreateMemoInput{
		UserID:        7,
		Content:       "two",
		AttachmentSet: "album-1",
	})
	if err != nil {
		t.Fatalf("second CreateMemo returned error: %v", err)
	}

	if createCalls != 1 {
		t.Fatalf("expected one backend call, got %d", createCalls)
	}
	if first != second {
		t.Fatal("expected cached memo pointer to be reused")
	}
}

func TestServiceCreateMemoDeletesExpiredMediaGroupCacheEntries(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		createMemo: func(context.Context, string, string) (*domain.Memo, error) {
			return &domain.Memo{Name: "memos/1"}, nil
		},
	}, store, "data.txt", nil, nil)

	service.mediaGroupCache.Store("old-album", mediaGroupCacheEntry{
		memo:      &domain.Memo{Name: "memos/old"},
		expiresAt: time.Now().Add(-time.Minute),
	})

	_, err := service.CreateMemo(context.Background(), CreateMemoInput{
		UserID:        7,
		Content:       "new",
		AttachmentSet: "new-album",
	})
	if err != nil {
		t.Fatalf("CreateMemo returned error: %v", err)
	}

	if _, ok := service.mediaGroupCache.Load("old-album"); ok {
		t.Fatal("expected expired media group cache entry to be deleted")
	}
}

func TestServiceAttachFileRequiresLinkedAccount(t *testing.T) {
	service := NewService(fakeBackend{baseURL: "https://example.test"}, &fakeTokenStore{}, "data.txt", nil, nil)

	err := service.AttachFile(context.Background(), 10, "memos/1", domain.FilePayload{})
	if !errors.Is(err, domain.ErrAccountNotLinked) {
		t.Fatalf("expected ErrAccountNotLinked, got %v", err)
	}
}

func TestServiceSearchMemosIncludesCreatorID(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	var gotCreatorID *int64
	var gotQuery string
	var gotLimit int

	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getCurrentUser: func(context.Context, string) (*domain.User, error) {
			return &domain.User{Name: "users/123"}, nil
		},
		searchMemos: func(_ context.Context, accessToken string, query string, creatorID *int64, limit int) ([]domain.Memo, error) {
			if accessToken != "token" {
				t.Fatalf("expected access token, got %q", accessToken)
			}
			gotCreatorID = creatorID
			gotQuery = query
			gotLimit = limit
			return []domain.Memo{{Name: "memos/1"}}, nil
		},
	}, store, "data.txt", nil, nil)

	memos, err := service.SearchMemos(context.Background(), 7, "needle", 3)
	if err != nil {
		t.Fatalf("SearchMemos returned error: %v", err)
	}
	if len(memos) != 1 {
		t.Fatalf("expected one memo, got %d", len(memos))
	}
	if gotCreatorID == nil || *gotCreatorID != 123 {
		t.Fatalf("expected creatorID 123, got %#v", gotCreatorID)
	}
	if gotQuery != "needle" || gotLimit != 3 {
		t.Fatalf("unexpected search args: query=%q limit=%d", gotQuery, gotLimit)
	}
}

func TestServiceSearchMemosInvalidToken(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getCurrentUser: func(context.Context, string) (*domain.User, error) {
			return nil, errors.New("unauthorized")
		},
	}, store, "data.txt", nil, nil)

	_, err := service.SearchMemos(context.Background(), 7, "needle", 3)
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceIsUserAdmin(t *testing.T) {
	service := NewService(fakeBackend{baseURL: "https://example.test"}, &fakeTokenStore{}, "data.txt", nil, []string{"Admin", " Root "})

	if !service.IsUserAdmin("admin") {
		t.Fatal("expected admin to be allowed")
	}
	if !service.IsUserAdmin("ROOT") {
		t.Fatal("expected root to be allowed")
	}
	if service.IsUserAdmin("") {
		t.Fatal("expected empty username not to be admin")
	}
	if service.IsUserAdmin("alice") {
		t.Fatal("expected alice not to be admin")
	}
}

func TestServiceGetStatusIncludesLinkedAccountData(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getCurrentUser: func(context.Context, string) (*domain.User, error) {
			return &domain.User{DisplayName: "Display Name"}, nil
		},
	}, store, "/tmp/data.txt", []string{"alice"}, []string{"admin"})

	report := service.GetStatus(context.Background(), 7)

	if !report.AccountLinked || !report.AccountTokenValid {
		t.Fatalf("expected linked valid account, got %+v", report)
	}
	if report.AccountDisplayName != "Display Name" {
		t.Fatalf("expected display name, got %q", report.AccountDisplayName)
	}
}

func TestServiceGetHealthIncludesAdminAndStoreData(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getInstanceProfile: func(context.Context) (*domain.InstanceProfile, error) {
			return &domain.InstanceProfile{InstanceURL: "https://memos.example.test"}, nil
		},
	}, store, "/tmp/data.txt", []string{"alice"}, []string{"admin"})

	service.Start(context.Background())
	report := service.GetHealth(context.Background())

	if !report.BackendAvailable {
		t.Fatalf("expected backend to be available, got error %q", report.BackendError)
	}
	if report.ServerURL != "https://example.test" {
		t.Fatalf("expected server URL, got %q", report.ServerURL)
	}
	if report.DataFile != "/tmp/data.txt" {
		t.Fatalf("expected data file, got %q", report.DataFile)
	}
	if report.InstanceURL != "https://memos.example.test" {
		t.Fatalf("expected instance URL to be populated, got %q", report.InstanceURL)
	}
	if report.LinkedUsers != 1 {
		t.Fatalf("expected one linked user, got %d", report.LinkedUsers)
	}
	if report.AllowedUsernames != 1 {
		t.Fatalf("expected one allowed username, got %d", report.AllowedUsernames)
	}
	if report.AdminUsernames != 1 {
		t.Fatalf("expected one admin username, got %d", report.AdminUsernames)
	}
}

func TestServiceUpdateMemoActionPinAndDelete(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	getMemoCalls := 0
	updateCalls := 0
	deleteCalls := 0
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getMemo: func(context.Context, string, string) (*domain.Memo, error) {
			getMemoCalls++
			return &domain.Memo{Name: "memos/1", Visibility: domain.VisibilityPrivate, Pinned: false}, nil
		},
		updateMemo: func(_ context.Context, _ string, memo *domain.Memo) (*domain.Memo, error) {
			updateCalls++
			if !memo.Pinned {
				t.Fatal("expected pin action to toggle pinned state")
			}
			return memo, nil
		},
		deleteMemo: func(context.Context, string, string) error {
			deleteCalls++
			return nil
		},
	}, store, "data.txt", nil, nil)

	updated, deleted, err := service.UpdateMemoAction(context.Background(), 7, ActionPin, "memos/1")
	if err != nil {
		t.Fatalf("pin action returned error: %v", err)
	}
	if deleted {
		t.Fatal("expected pin action not to delete")
	}
	if updated == nil || !updated.Pinned {
		t.Fatalf("expected updated pinned memo, got %#v", updated)
	}

	updated, deleted, err = service.UpdateMemoAction(context.Background(), 7, ActionDelete, "memos/1")
	if err != nil {
		t.Fatalf("delete action returned error: %v", err)
	}
	if !deleted || updated != nil {
		t.Fatalf("expected delete result, got deleted=%v updated=%#v", deleted, updated)
	}
	if getMemoCalls != 1 || updateCalls != 1 || deleteCalls != 1 {
		t.Fatalf("unexpected backend calls: get=%d update=%d delete=%d", getMemoCalls, updateCalls, deleteCalls)
	}
}

func TestServiceUpdateMemoActionUnknown(t *testing.T) {
	store := &fakeTokenStore{tokens: map[int64]string{7: "token"}}
	service := NewService(fakeBackend{
		baseURL: "https://example.test",
		getMemo: func(context.Context, string, string) (*domain.Memo, error) {
			return &domain.Memo{Name: "memos/1"}, nil
		},
	}, store, "data.txt", nil, nil)

	_, _, err := service.UpdateMemoAction(context.Background(), 7, MemoAction("bad"), "memos/1")
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestServiceMemoBaseURLFallsBackToBackendURL(t *testing.T) {
	service := NewService(fakeBackend{baseURL: "https://example.test"}, &fakeTokenStore{}, "data.txt", nil, nil)
	if got := service.MemoBaseURL(); got != "https://example.test" {
		t.Fatalf("expected backend base URL fallback, got %q", got)
	}
}

func TestDisplayNameOf(t *testing.T) {
	tests := []struct {
		name string
		user *domain.User
		want string
	}{
		{name: "nil", user: nil, want: ""},
		{name: "display name", user: &domain.User{DisplayName: "Display"}, want: "Display"},
		{name: "username fallback", user: &domain.User{Username: "user"}, want: "user"},
		{name: "name fallback", user: &domain.User{Name: "Name"}, want: "Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayNameOf(tt.user); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
