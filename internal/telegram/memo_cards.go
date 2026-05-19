package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

const callbackSearchMore = "searchmore"

func (t *Bot) sendMemoCard(ctx context.Context, chatID int64, memo domain.Memo, replyToMessageID int) {
	params := t.memoCardParams(chatID, memo)
	params.DisableNotification = true
	if replyToMessageID != 0 {
		params.ReplyParameters = &models.ReplyParameters{MessageID: replyToMessageID}
	}
	t.bot.SendMessage(ctx, params)
}

func (t *Bot) sendMemoUpdateCard(ctx context.Context, chatID int64, memo domain.Memo, replyToMessageID int) {
	params := t.memoCardParams(chatID, memo)
	params.Text = formatMemoUpdatedCard(memo, t.service.MemoBaseURL())
	if replyToMessageID != 0 {
		params.ReplyParameters = &models.ReplyParameters{MessageID: replyToMessageID}
	}
	t.bot.SendMessage(ctx, params)
}

func (t *Bot) memoCardParams(chatID int64, memo domain.Memo) *bot.SendMessageParams {
	return &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        formatMemoCard(memo, t.service.MemoBaseURL()),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard(&memo),
	}
}

func (t *Bot) sendSearchMorePrompt(ctx context.Context, chatID int64, query string, sessionID string) {
	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("More results are available for %q.", query),
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "More Results", CallbackData: fmt.Sprintf("%s %s", callbackSearchMore, sessionID)},
				},
			},
		},
	})
}

func formatMemoCard(memo domain.Memo, baseURL string) string {
	return fmt.Sprintf(
		"*%s* [%s](%s)%s\n%s",
		escapeMarkdownV2(string(memo.Visibility)),
		escapeMarkdownV2(memo.Name),
		escapeMarkdownV2URL(memoURL(baseURL, memo.Name)),
		pinnedSuffix(memo.Pinned),
		escapeMarkdownV2(memoPreview(memo.Content)),
	)
}

func formatMemoUpdatedCard(memo domain.Memo, baseURL string) string {
	return "Memo updated\n" + formatMemoCard(memo, baseURL)
}

func memoURL(baseURL string, memoName string) string {
	memoUID, err := domain.ExtractMemoUIDFromName(memoName)
	if err != nil {
		return strings.TrimRight(baseURL, "/")
	}
	return strings.TrimRight(baseURL, "/") + "/memos/" + memoUID
}

func memoPreview(content string) string {
	const maxLength = 700
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "(empty memo)"
	}
	runes := []rune(trimmed)
	if len(runes) <= maxLength {
		return trimmed
	}
	return string(runes[:maxLength]) + "..."
}

func pinnedSuffix(pinned bool) string {
	if !pinned {
		return ""
	}
	return " " + escapeMarkdownV2("📌")
}
