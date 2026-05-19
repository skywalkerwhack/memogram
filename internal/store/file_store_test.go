package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	userID, token := parseLine("123:abc:def")
	if userID != 123 {
		t.Fatalf("expected userID 123, got %d", userID)
	}
	if token != "abc:def" {
		t.Fatalf("expected token with colon, got %q", token)
	}
}

func TestSaveAndLoadUserAccessTokens(t *testing.T) {
	dataPath := filepath.Join(t.TempDir(), "data.txt")

	store, err := NewFileTokenStore(dataPath)
	if err != nil {
		t.Fatalf("init store: %v", err)
	}

	if err := store.SetUserAccessToken(42, "token-one"); err != nil {
		t.Fatalf("set token-one: %v", err)
	}
	if err := store.SetUserAccessToken(7, "token:two"); err != nil {
		t.Fatalf("set token:two: %v", err)
	}

	reloaded, err := NewFileTokenStore(dataPath)
	if err != nil {
		t.Fatalf("init reloaded store: %v", err)
	}

	token, ok := reloaded.GetUserAccessToken(42)
	if !ok || token != "token-one" {
		t.Fatalf("expected token-one for user 42, got %q", token)
	}

	token, ok = reloaded.GetUserAccessToken(7)
	if !ok || token != "token:two" {
		t.Fatalf("expected token:two for user 7, got %q", token)
	}
}

func TestDeleteUserAccessToken(t *testing.T) {
	dataPath := filepath.Join(t.TempDir(), "data.txt")

	store, err := NewFileTokenStore(dataPath)
	if err != nil {
		t.Fatalf("init store: %v", err)
	}

	if err := store.SetUserAccessToken(42, "token-one"); err != nil {
		t.Fatalf("set token-one: %v", err)
	}
	if err := store.SetUserAccessToken(7, "token-two"); err != nil {
		t.Fatalf("set token-two: %v", err)
	}

	ok, err := store.DeleteUserAccessToken(42)
	if err != nil {
		t.Fatalf("delete user 42: %v", err)
	}
	if !ok {
		t.Fatal("expected delete to report existing user")
	}
	if _, ok := store.GetUserAccessToken(42); ok {
		t.Fatal("expected user 42 token to be removed from cache")
	}

	reloaded, err := NewFileTokenStore(dataPath)
	if err != nil {
		t.Fatalf("init reloaded store: %v", err)
	}

	if _, ok := reloaded.GetUserAccessToken(42); ok {
		t.Fatal("expected user 42 token to be removed from file")
	}
	token, ok := reloaded.GetUserAccessToken(7)
	if !ok || token != "token-two" {
		t.Fatalf("expected remaining token for user 7, got %q", token)
	}

	ok, err = store.DeleteUserAccessToken(999)
	if err != nil {
		t.Fatalf("delete user 999: %v", err)
	}
	if ok {
		t.Fatal("expected delete to report missing user")
	}
}

func TestSetUserAccessTokenRollsBackOnSaveFailure(t *testing.T) {
	dataPath := filepath.Join(t.TempDir(), "data.txt")

	store, err := NewFileTokenStore(dataPath)
	if err != nil {
		t.Fatalf("init store: %v", err)
	}
	if err := store.SetUserAccessToken(42, "token-one"); err != nil {
		t.Fatalf("set initial token: %v", err)
	}

	blockingDir := filepath.Join(t.TempDir(), "blocked")
	if err := os.MkdirAll(blockingDir, 0o755); err != nil {
		t.Fatalf("mkdir blocked dir: %v", err)
	}
	store.dataPath = blockingDir

	err = store.SetUserAccessToken(42, "token-two")
	if err == nil {
		t.Fatal("expected save failure")
	}

	token, ok := store.GetUserAccessToken(42)
	if !ok || token != "token-one" {
		t.Fatalf("expected rollback to token-one, got %q", token)
	}
}

func TestDeleteUserAccessTokenRollsBackOnSaveFailure(t *testing.T) {
	dataPath := filepath.Join(t.TempDir(), "data.txt")

	store, err := NewFileTokenStore(dataPath)
	if err != nil {
		t.Fatalf("init store: %v", err)
	}
	if err := store.SetUserAccessToken(42, "token-one"); err != nil {
		t.Fatalf("set initial token: %v", err)
	}

	blockingDir := filepath.Join(t.TempDir(), "blocked")
	if err := os.MkdirAll(blockingDir, 0o755); err != nil {
		t.Fatalf("mkdir blocked dir: %v", err)
	}
	store.dataPath = blockingDir

	ok, err := store.DeleteUserAccessToken(42)
	if !ok {
		t.Fatal("expected delete to report existing user")
	}
	if err == nil {
		t.Fatal("expected save failure")
	}

	token, stillPresent := store.GetUserAccessToken(42)
	if !stillPresent || token != "token-one" {
		t.Fatalf("expected rollback to keep token-one, got %q", token)
	}
}

func TestDeleteUserAccessTokenMissingWithoutError(t *testing.T) {
	store := &FileTokenStore{}

	ok, err := store.DeleteUserAccessToken(99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected missing user to report false")
	}
}
