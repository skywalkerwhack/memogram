package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain host", in: "localhost:5230", want: "http://localhost:5230"},
		{name: "dns prefix", in: "dns:localhost:5230", want: "http://localhost:5230"},
		{name: "https untouched", in: "https://example.test", want: "https://example.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeBaseURL(tt.in); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestParseAllowedUsernames(t *testing.T) {
	got := parseAllowedUsernames(" Alice,BOB, , carol ")
	want := []string{"alice", "bob", "carol"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	restoreCWD(t, tempDir)
	unsetEnv(t, "SERVER_ADDR")
	unsetEnv(t, "BOT_TOKEN")
	unsetEnv(t, "BOT_PROXY_ADDR")
	unsetEnv(t, "DATA")
	unsetEnv(t, "ALLOWED_USERNAMES")

	t.Setenv("SERVER_ADDR", "dns:localhost:5230")
	t.Setenv("BOT_TOKEN", "bot-token")
	t.Setenv("BOT_PROXY_ADDR", "http://proxy.test")
	t.Setenv("DATA", "tokens.txt")
	t.Setenv("ALLOWED_USERNAMES", "Alice,bob")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ServerAddr != "http://localhost:5230" {
		t.Fatalf("unexpected server addr %q", cfg.ServerAddr)
	}
	if cfg.BotToken != "bot-token" {
		t.Fatalf("unexpected bot token %q", cfg.BotToken)
	}
	if cfg.BotProxyAddr != "http://proxy.test" {
		t.Fatalf("unexpected proxy addr %q", cfg.BotProxyAddr)
	}
	if !strings.HasSuffix(cfg.Data, string(filepath.Separator)+"tokens.txt") {
		t.Fatalf("expected absolute data path, got %q", cfg.Data)
	}
	if want := []string{"alice", "bob"}; !reflect.DeepEqual(cfg.AllowedUsernames, want) {
		t.Fatalf("expected usernames %v, got %v", want, cfg.AllowedUsernames)
	}
}

func TestLoadFromDotEnvAndDefaultDataPath(t *testing.T) {
	tempDir := t.TempDir()
	restoreCWD(t, tempDir)
	unsetEnv(t, "SERVER_ADDR")
	unsetEnv(t, "BOT_TOKEN")
	unsetEnv(t, "BOT_PROXY_ADDR")
	unsetEnv(t, "DATA")
	unsetEnv(t, "ALLOWED_USERNAMES")

	envContent := "SERVER_ADDR=https://memos.example.test\nBOT_TOKEN=dotenv-token\nALLOWED_USERNAMES=alice\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ServerAddr != "https://memos.example.test" {
		t.Fatalf("unexpected server addr %q", cfg.ServerAddr)
	}
	if cfg.BotToken != "dotenv-token" {
		t.Fatalf("unexpected bot token %q", cfg.BotToken)
	}
	if !strings.HasSuffix(cfg.Data, string(filepath.Separator)+"data.txt") {
		t.Fatalf("expected default data.txt path, got %q", cfg.Data)
	}
}

func TestLoadRejectsDirectoryDataPath(t *testing.T) {
	tempDir := t.TempDir()
	restoreCWD(t, tempDir)
	unsetEnv(t, "SERVER_ADDR")
	unsetEnv(t, "BOT_TOKEN")
	unsetEnv(t, "BOT_PROXY_ADDR")
	unsetEnv(t, "DATA")
	unsetEnv(t, "ALLOWED_USERNAMES")

	dataDir := filepath.Join(tempDir, "store")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir data dir: %v", err)
	}

	t.Setenv("SERVER_ADDR", "localhost:5230")
	t.Setenv("BOT_TOKEN", "bot-token")
	t.Setenv("DATA", dataDir)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for directory data path")
	}
	if !strings.Contains(err.Error(), "cannot be a directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func restoreCWD(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previous)
	})
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	value, ok := os.LookupEnv(key)
	if ok {
		t.Cleanup(func() {
			_ = os.Setenv(key, value)
		})
	} else {
		t.Cleanup(func() {
			_ = os.Unsetenv(key)
		})
	}
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
}
