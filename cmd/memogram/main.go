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

	"github.com/usememos/memogram/internal/app"
	"github.com/usememos/memogram/internal/config"
	"github.com/usememos/memogram/internal/memos"
	"github.com/usememos/memogram/internal/store"
	"github.com/usememos/memogram/internal/telegram"
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

	backend := memos.NewBackend(cfg.ServerAddr, http.DefaultClient)
	service := app.NewService(backend, tokenStore, cfg.Data, cfg.AllowedUsernames)

	tgBot, err := telegram.NewBot(cfg, service)
	if err != nil {
		log.Fatal(err)
	}

	tgBot.Start(ctx)
}
