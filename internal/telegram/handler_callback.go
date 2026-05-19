package telegram

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (t *Bot) handleCallbackQuery(ctx context.Context, b *bot.Bot, update *models.Update) {
	callbackData := update.CallbackQuery.Data
	parts := splitCallbackData(callbackData)
	if len(parts) != 2 {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Invalid command",
			ShowAlert:       true,
		})
		return
	}

	command := parts[0]
	argument := parts[1]
	switch command {
	case callbackSearchMore:
		t.handleSearchMoreCallback(ctx, b, update, argument)
		return
	case "edit":
		t.handleEditCallback(ctx, b, update, argument)
		return
	}

	action, memoName := app.MemoAction(command), argument
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

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        formatMemoUpdatedCard(*memo, t.service.MemoBaseURL()),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard(memo),
	})

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Memo updated",
	})
}

func (t *Bot) handleEditCallback(ctx context.Context, b *bot.Bot, update *models.Update, memoName string) {
	memo, err := t.service.BeginMemoEdit(ctx, update.CallbackQuery.From.ID, memoName)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotLinked) {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Please start the bot with /start <access_token>",
				ShowAlert:       true,
			})
			return
		}

		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Failed to start editing this memo",
			ShowAlert:       true,
		})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Send the replacement text for this memo.",
		ShowAlert:       true,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text: fmt.Sprintf(
			"Editing [%s](%s)\n\nCurrent content:\n%s\n\nSend the replacement text, or /cancel.",
			escapeMarkdownV2(memo.Name),
			escapeMarkdownV2URL(memoURL(t.service.MemoBaseURL(), memo.Name)),
			escapeMarkdownV2(memoPreview(memo.Content)),
		),
		ParseMode: models.ParseModeMarkdown,
	})
}

func (t *Bot) handleSearchMoreCallback(ctx context.Context, b *bot.Bot, update *models.Update, sessionID string) {
	userID, query, offset, limit, ok := t.service.LoadSearchSession(sessionID)
	if !ok || userID != update.CallbackQuery.From.ID {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "This search session is no longer available.",
			ShowAlert:       true,
		})
		return
	}

	page, err := t.service.SearchMemosPage(ctx, userID, query, offset, limit)
	if err != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Failed to load more results",
			ShowAlert:       true,
		})
		return
	}

	if len(page.Memos) == 0 {
		t.service.DeleteSearchSession(sessionID)
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      fmt.Sprintf("No more results for %q.", query),
		})
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "No more results",
		})
		return
	}

	for _, memo := range page.Memos {
		t.sendMemoCard(ctx, update.CallbackQuery.Message.Message.Chat.ID, memo, 0)
	}

	if page.HasMore {
		t.service.AdvanceSearchSession(sessionID, page.Offset+len(page.Memos))
	} else {
		t.service.DeleteSearchSession(sessionID)
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      fmt.Sprintf("All results loaded for %q.", query),
		})
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Loaded more results",
	})
}

func failureText(action app.MemoAction) string {
	if action == app.ActionDelete {
		return "Failed to delete memo"
	}
	return "Failed to update memo"
}

func splitCallbackData(callbackData string) []string {
	for i := 0; i < len(callbackData); i++ {
		if callbackData[i] == ' ' {
			return []string{callbackData[:i], callbackData[i+1:]}
		}
	}
	return nil
}
