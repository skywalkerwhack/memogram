package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (t *Bot) handleCallbackQuery(ctx context.Context, b *bot.Bot, update *models.Update) {
	callbackData := update.CallbackQuery.Data
	parts := strings.SplitN(callbackData, " ", 2)
	if len(parts) != 2 {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Invalid command",
			ShowAlert:       true,
		})
		return
	}

	action, memoName := app.MemoAction(parts[0]), parts[1]
	memo, deleted, err := t.service.UpdateMemoAction(ctx, update.CallbackQuery.From.ID, action, memoName)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountNotLinked):
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Please start the bot with /start <access_token>",
				ShowAlert:       true,
			})
		default:
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            failureText(action),
				ShowAlert:       true,
			})
		}
		return
	}

	if deleted {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      fmt.Sprintf("Memo deleted: %s", memoName),
		})
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Memo deleted",
		})
		return
	}

	memoUID, err := domain.ExtractMemoUIDFromName(memo.Name)
	if err != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Failed to update memo",
		})
		return
	}

	pinnedMarker := ""
	if memo.Pinned {
		pinnedMarker = "📌"
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        formatMemoUpdatedMessage(memo.Visibility, memo.Name, t.service.MemoBaseURL(), memoUID, pinnedMarker),
		ParseMode:   telegramMarkdownParseMode,
		ReplyMarkup: keyboard(memo),
	})

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Memo updated",
	})
}

func formatMemoUpdatedMessage(visibility domain.Visibility, memoName string, baseURL string, memoUID string, pinnedMarker string) string {
	message := fmt.Sprintf(
		"Memo updated as *%s* with [%s](%s)",
		escapeMarkdownV2(string(visibility)),
		escapeMarkdownV2(memoName),
		escapeMarkdownV2URL(strings.TrimRight(baseURL, "/")+"/memos/"+memoUID),
	)
	if pinnedMarker != "" {
		message += " " + escapeMarkdownV2(pinnedMarker)
	}
	return message
}

func failureText(action app.MemoAction) string {
	if action == app.ActionDelete {
		return "Failed to delete memo"
	}
	return "Failed to update memo"
}
