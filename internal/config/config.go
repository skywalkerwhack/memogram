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
	ServerAddr       string `env:"SERVER_ADDR,required"`
	BotToken         string `env:"BOT_TOKEN,required"`
	BotProxyAddr     string `env:"BOT_PROXY_ADDR"`
	Data             string `env:"DATA"`
	AllowedUsernames string `env:"ALLOWED_USERNAMES"`
}

type Config struct {
	ServerAddr       string
	BotToken         string
	BotProxyAddr     string
	Data             string
	AllowedUsernames []string
}

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

	return &Config{
		ServerAddr:       normalizeBaseURL(raw.ServerAddr),
		BotToken:         raw.BotToken,
		BotProxyAddr:     raw.BotProxyAddr,
		Data:             absDataPath,
		AllowedUsernames: parseAllowedUsernames(raw.AllowedUsernames),
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
