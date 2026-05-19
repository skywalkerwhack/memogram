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

package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/config"
	"github.com/skywalkerwhack/memogram/internal/memos"
	"github.com/skywalkerwhack/memogram/internal/store"
	"github.com/skywalkerwhack/memogram/internal/telegram"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	tokenStore, err := store.NewFileTokenStore(cfg.Data)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{
		Timeout: 35 * time.Second,
	}

	backend := memos.NewBackend(cfg.ServerAddr, httpClient)
	service := app.NewService(backend, tokenStore, cfg.Data, cfg.AllowedUsernames, cfg.AdminUsernames)

	tgBot, err := telegram.NewBot(cfg, service, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	tgBot.Start(ctx)
}
