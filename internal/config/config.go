package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type rawConfig struct {
	ServerAddr         string `env:"SERVER_ADDR,required"`
	BotToken           string `env:"BOT_TOKEN,required"`
	BotProxyAddr       string `env:"BOT_PROXY_ADDR"`
	Data               string `env:"DATA"`
	AllowedUsernames   string `env:"ALLOWED_USERNAMES"`
	AdminUsernames     string `env:"ADMIN_USERNAMES"`
	MaxAttachmentBytes int64  `env:"MAX_ATTACHMENT_BYTES"`
}

type Config struct {
	ServerAddr         string
	BotToken           string
	BotProxyAddr       string
	Data               string
	AllowedUsernames   []string
	AdminUsernames     []string
	MaxAttachmentBytes int64
}

const DefaultMaxAttachmentBytes int64 = 20 * 1024 * 1024

func Load() (*Config, error) {
	const envFileName = ".env"
	if _, err := os.Stat(envFileName); err == nil {
		if err := godotenv.Load(envFileName); err != nil {
			return nil, fmt.Errorf("load %s: %w", envFileName, err)
		}
	}

	var raw rawConfig
	if err := env.Parse(&raw); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	dataPath := raw.Data
	if dataPath == "" {
		dataPath = "data.txt"
	}

	absDataPath, err := filepath.Abs(dataPath)
	if err != nil {
		return nil, fmt.Errorf("resolve data path %s: %w", dataPath, err)
	}

	if info, err := os.Stat(absDataPath); err == nil && info.IsDir() {
		return nil, fmt.Errorf("config file cannot be a directory: %s", absDataPath)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("access config file %s: %w", absDataPath, err)
	}

	maxAttachmentBytes := raw.MaxAttachmentBytes
	if maxAttachmentBytes == 0 {
		maxAttachmentBytes = DefaultMaxAttachmentBytes
	}
	if maxAttachmentBytes < 0 {
		return nil, fmt.Errorf("MAX_ATTACHMENT_BYTES must be greater than or equal to 0")
	}

	return &Config{
		ServerAddr:         normalizeBaseURL(raw.ServerAddr),
		BotToken:           raw.BotToken,
		BotProxyAddr:       raw.BotProxyAddr,
		Data:               absDataPath,
		AllowedUsernames:   parseAllowedUsernames(raw.AllowedUsernames),
		AdminUsernames:     parseAllowedUsernames(raw.AdminUsernames),
		MaxAttachmentBytes: maxAttachmentBytes,
	}, nil
}

func normalizeBaseURL(addr string) string {
	baseURL := strings.TrimPrefix(addr, "dns:")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func parseAllowedUsernames(raw string) []string {
	parts := strings.Split(raw, ",")
	allowed := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.ToLower(strings.TrimSpace(part))
		if trimmed == "" {
			continue
		}
		allowed = append(allowed, trimmed)
	}
	return allowed
}
