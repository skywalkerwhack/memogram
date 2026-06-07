package telegram

import (
	"fmt"

	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

const (
	callbackDeletePrompt  = "delete_prompt"
	callbackDeleteConfirm = "delete_confirm"
	callbackDeleteCancel  = "delete_cancel"
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
				{Text: "Delete", CallbackData: fmt.Sprintf("%s %s", callbackDeletePrompt, memo.Name)},
			},
		},
	}
}

func deleteConfirmationKeyboard(memoName string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Confirm delete", CallbackData: fmt.Sprintf("%s %s", callbackDeleteConfirm, memoName)},
				{Text: "Cancel", CallbackData: fmt.Sprintf("%s %s", callbackDeleteCancel, memoName)},
			},
		},
	}
}
