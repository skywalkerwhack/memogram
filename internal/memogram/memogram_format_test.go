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
	"testing"

	"github.com/go-telegram/bot/models"
)

func TestFormatContent_MixedEntities(t *testing.T) {
	content := "See example.com and bold text link"
	entities := []models.MessageEntity{
		{
			Type:   models.MessageEntityTypeURL,
			Offset: 4,
			Length: 11,
		},
		{
			Type:   models.MessageEntityTypeBold,
			Offset: 20,
			Length: 4,
		},
		{
			Type:   models.MessageEntityTypeTextLink,
			Offset: 30,
			Length: 4,
			URL:    "https://example.com",
		},
	}

	got := formatContent(content, entities)
	want := "See [example.com](example.com) and **bold** text [link](https://example.com)"
	if got != want {
		t.Fatalf("unexpected content:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestFormatContent_OutOfOrderEntities(t *testing.T) {
	content := "Italic and bold"
	entities := []models.MessageEntity{
		{
			Type:   models.MessageEntityTypeBold,
			Offset: 11,
			Length: 4,
		},
		{
			Type:   models.MessageEntityTypeItalic,
			Offset: 0,
			Length: 6,
		},
	}

	got := formatContent(content, entities)
	want := "*Italic* and **bold**"
	if got != want {
		t.Fatalf("unexpected content:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestFormatContent_OverlappingEntities(t *testing.T) {
	content := "Overlap test"
	entities := []models.MessageEntity{
		{
			Type:   models.MessageEntityTypeBold,
			Offset: 0,
			Length: 7,
		},
		{
			Type:   models.MessageEntityTypeItalic,
			Offset: 5,
			Length: 4,
		},
	}

	got := formatContent(content, entities)
	want := "**Overlap** test"
	if got != want {
		t.Fatalf("unexpected content:\nwant: %q\ngot:  %q", want, got)
	}
}
