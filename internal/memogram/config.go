// memogram is a Telegram bot for saving messages into a Memos instance.
// Copyright (C) 2026  skywalkerwhack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package memogram

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddr       string `env:"SERVER_ADDR,required"`
	BotToken         string `env:"BOT_TOKEN,required"`
	BotProxyAddr     string `env:"BOT_PROXY_ADDR"`
	Data             string `env:"DATA"`
	AllowedUsernames string `env:"ALLOWED_USERNAMES"`
}

func getConfigFromEnv() (*Config, error) {
	envFileName := ".env"
	if _, err := os.Stat(envFileName); err == nil {
		if err := godotenv.Load(envFileName); err != nil {
			return nil, fmt.Errorf("load %s: %w", envFileName, err)
		}
	}

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	if config.Data == "" {
		// Default to `data.txt` if not specified.
		config.Data = "data.txt"
	}

	fileInfo, err := os.Stat(config.Data)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the file with default permissions
			file, err := os.OpenFile(config.Data, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to create config file %s: %w", config.Data, err)
			}
			file.Close()

			// Get file info after creating the file
			fileInfo, err = os.Stat(config.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to get file info after creating %s: %w", config.Data, err)
			}
		} else {
			return nil, fmt.Errorf("failed to access config file %s: %w", config.Data, err)
		}
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("config file cannot be a directory: %s", config.Data)
	}

	config.Data, err = filepath.Abs(config.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for config file %s: %w", config.Data, err)
	}

	return &config, nil
}
