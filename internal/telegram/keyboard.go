package telegram

import (
	"fmt"

	"github.com/go-telegram/bot/models"
	"github.com/usememos/memogram/internal/domain"
)

func keyboard(memo *domain.Memo) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Public", CallbackData: fmt.Sprintf("public %s", memo.Name)},
				{Text: "Private", CallbackData: fmt.Sprintf("private %s", memo.Name)},
				{Text: "Pin", CallbackData: fmt.Sprintf("pin %s", memo.Name)},
			},
			{
				{Text: "Delete", CallbackData: fmt.Sprintf("delete %s", memo.Name)},
			},
		},
	}
}
