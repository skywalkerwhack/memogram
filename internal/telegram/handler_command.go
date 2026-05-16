package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

const (
	commandStart  = "/start"
	commandHelp   = "/help"
	commandUnlink = "/unlink"
	commandSearch = "/search"
	commandStatus = "/status"
	commandPing   = "/ping"
)

func (t *Bot) handleCommand(ctx context.Context, update *models.Update) bool {
	message := update.Message

	switch {
	case strings.HasPrefix(message.Text, commandStart+" ") || message.Text == commandStart:
		t.startHandler(ctx, update)
	case strings.HasPrefix(message.Text, commandHelp+" ") || message.Text == commandHelp:
		t.helpHandler(ctx, update)
	case strings.HasPrefix(message.Text, commandUnlink+" ") || message.Text == commandUnlink:
		t.unlinkHandler(ctx, update)
	case strings.HasPrefix(message.Text, commandSearch+" ") || message.Text == commandSearch:
		t.searchHandler(ctx, update)
	case strings.HasPrefix(message.Text, commandStatus+" ") || message.Text == commandStatus:
		t.statusHandler(ctx, update)
	case strings.HasPrefix(message.Text, commandPing+" ") || message.Text == commandPing:
		t.pingHandler(ctx, update)
	default:
		return false
	}
	return true
}

func (t *Bot) startHandler(ctx context.Context, update *models.Update) {
	accessToken := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, commandStart))
	if accessToken == "" {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /start <access_token>",
		})
		return
	}

	displayName, err := t.service.LinkAccount(ctx, update.Message.From.ID, accessToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Invalid access token",
			})
			return
		}
		t.sendError(update.Message.Chat.ID, err)
		return
	}

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Hello %s!", displayName),
	})
}

func (t *Bot) helpHandler(ctx context.Context, update *models.Update) {
	lines := []string{
		"Memogram commands",
		"/start <access_token> - link this Telegram account to Memos",
		"/unlink - remove the saved Memos token for this Telegram account",
		"/search <words> - search your saved memos",
		"/status - show bot and account status",
		"/ping - check whether the bot is alive",
		"/help - show this help message",
		"",
		"Send text, photos, voice messages, videos, or documents to save them as memos.",
	}

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   strings.Join(lines, "\n"),
	})
}

func (t *Bot) unlinkHandler(ctx context.Context, update *models.Update) {
	deleted, err := t.service.UnlinkAccount(update.Message.From.ID)
	if err != nil {
		t.sendError(update.Message.Chat.ID, err)
		return
	}
	if !deleted {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No linked Memos account was found for this Telegram account.",
		})
		return
	}

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Your Memos account has been disconnected. Use /start <access_token> to link it again.",
	})
}

func (t *Bot) searchHandler(ctx context.Context, update *models.Update) {
	searchString := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, commandSearch))
	if searchString == "" {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /search <words>",
		})
		return
	}

	memos, err := t.service.SearchMemos(ctx, update.Message.From.ID, searchString, 10)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountNotLinked):
			t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Please start the bot with /start <access_token>",
			})
		case errors.Is(err, domain.ErrInvalidToken):
			t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Invalid access token",
			})
		default:
			t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Failed to search memos",
			})
		}
		return
	}

	if len(memos) == 0 {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No memos found for the specified search criteria.",
		})
		return
	}

	for _, memo := range memos {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   memo.Name + "\n" + memo.Content,
		})
	}
}

func (t *Bot) statusHandler(ctx context.Context, update *models.Update) {
	report := t.service.GetStatus(ctx, update.Message.From.ID)

	lines := []string{
		"Memogram status",
		fmt.Sprintf("Server: %s", report.ServerURL),
		fmt.Sprintf("Data file: %s", report.DataFile),
		backendStatusLine(report),
	}

	if report.InstanceURL != "" {
		lines = append(lines, fmt.Sprintf("Instance URL: %s", report.InstanceURL))
	}

	if report.AllowedUsernames == 0 {
		lines = append(lines, "Allowed usernames: unrestricted")
	} else {
		lines = append(lines, fmt.Sprintf("Allowed usernames: %d configured", report.AllowedUsernames))
	}

	lines = append(lines, fmt.Sprintf("Linked Telegram users: %d", report.LinkedUsers))

	switch {
	case !report.AccountLinked:
		lines = append(lines, "Account link: not connected")
		lines = append(lines, "Use /start <access_token> to connect this Telegram account.")
	case !report.AccountTokenValid:
		lines = append(lines, "Account link: saved token is invalid")
		lines = append(lines, "Run /start <access_token> again to refresh it.")
	default:
		displayName := report.AccountDisplayName
		if displayName == "" {
			displayName = "connected"
		}
		lines = append(lines, "Account link: connected as "+displayName)
	}

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   strings.Join(lines, "\n"),
	})
}

func backendStatusLine(report app.StatusReport) string {
	if !report.BackendAvailable {
		return fmt.Sprintf("Backend latency: unavailable (%s)", report.BackendError)
	}
	return fmt.Sprintf("Backend latency: %s", formatLatency(report.BackendLatency))
}

func formatLatency(latency time.Duration) string {
	switch {
	case latency < time.Millisecond:
		return fmt.Sprintf("%dµs", latency.Microseconds())
	case latency < time.Second:
		return latency.Round(time.Millisecond).String()
	default:
		return latency.Round(10 * time.Millisecond).String()
	}
}

func (t *Bot) pingHandler(ctx context.Context, update *models.Update) {
	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Pong!",
	})
}
