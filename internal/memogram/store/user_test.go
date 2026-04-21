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

package store

import (
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

	store := NewStore(dataPath)
	if err := store.Init(); err != nil {
		t.Fatalf("init store: %v", err)
	}

	store.SetUserAccessToken(42, "token-one")
	store.SetUserAccessToken(7, "token:two")

	reloaded := NewStore(dataPath)
	if err := reloaded.Init(); err != nil {
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
