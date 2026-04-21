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
	"fmt"
	"sync"
)

type Store struct {
	Data string

	userAccessTokenCache sync.Map // map[int64]string
}

func NewStore(data string) *Store {
	return &Store{
		Data: data,

		userAccessTokenCache: sync.Map{},
	}
}

func (s *Store) Init() error {
	if err := s.loadUserAccessTokenMapFromFile(); err != nil {
		return fmt.Errorf("failed to load user access token map from file: %w", err)
	}

	return nil
}
