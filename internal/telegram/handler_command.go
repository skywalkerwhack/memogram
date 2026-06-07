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
	commandStart   = "/start"
	commandHelp    = "/help"
	commandUnlink  = "/unlink"
	commandSearch  = "/search"
	commandAccount = "/account"
	commandMe      = "/me"
	commandPing    = "/ping"
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
	case strings.HasPrefix(message.Text, commandAccount+" ") || message.Text == commandAccount:
		t.accountHandler(ctx, update)
	case strings.HasPrefix(message.Text, commandMe+" ") || message.Text == commandMe:
		t.accountHandler(ctx, update)
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
			Text:   startUsageMessage(),
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

func startUsageMessage() string {
	return strings.Join([]string{
		"Connect Memogram to your Memos account",
		"",
		"1. Open your Memos account settings.",
		"2. Create an access token.",
		"3. Send it here like this:",
		"/start <access_token>",
		"",
		"Your token is stored locally by this bot and used only to save memos for your Telegram account.",
	}, "\n")
}

func (t *Bot) helpHandler(ctx context.Context, update *models.Update) {
	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   helpMessage(),
	})
}

func helpMessage() string {
	return strings.Join([]string{
		"Memogram help",
		"",
		"Save",
		"Send text, photos, voice messages, videos, or documents. I will save them to Memos.",
		"",
		"Search",
		"Use /search words to find saved memos.",
		"",
		"Account",
		"Use /account to check your connection.",
		"Use /start <access_token> to connect or refresh your token.",
		"Use /unlink to disconnect this Telegram account.",
		"",
		"Admin",
		"Use /ping to check backend diagnostics if you are an admin.",
		"",
		"Help",
		"Use /help to show this message again.",
	}, "\n")
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
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   memoSearchErrorMessage(err),
		})
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

func memoSearchErrorMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrAccountNotLinked):
		return accountNotLinkedMessage()
	case errors.Is(err, domain.ErrInvalidToken):
		return invalidTokenMessage()
	case errors.Is(err, domain.ErrBackendUnavailable):
		return memosUnavailableMessage()
	default:
		return memosUnavailableMessage()
	}
}

func accountNotLinkedMessage() string {
	return "Please connect your Memos account first with /start <access_token>."
}

func invalidTokenMessage() string {
	return "Your Memos access token no longer works. Send /start <access_token> to reconnect."
}

func memosUnavailableMessage() string {
	return "I could not reach your Memos server. Check that Memos is running, then try again."
}

func attachmentTooLargeMessage(maxBytes int64) string {
	if maxBytes <= 0 {
		return "That attachment is too large to process."
	}
	return fmt.Sprintf("That attachment is too large. The current limit is %s.", formatByteSize(maxBytes))
}

func formatByteSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes)
	for _, suffix := range []string{"KiB", "MiB", "GiB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f TiB", value/unit)
}

func (t *Bot) accountHandler(ctx context.Context, update *models.Update) {
	report := t.service.GetStatus(ctx, update.Message.From.ID)

	lines := []string{
		"Memogram account",
	}

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

func (t *Bot) pingHandler(ctx context.Context, update *models.Update) {
	username := update.Message.From.Username
	if !t.service.IsUserAdmin(username) {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ping diagnostics are only available to configured admins.",
		})
		return
	}

	report := t.service.GetHealth(ctx)
	lines := []string{
		"Memogram ping",
		fmt.Sprintf("Server: %s", report.ServerURL),
		fmt.Sprintf("Data file: %s", report.DataFile),
		backendHealthLine(report),
	}

	if report.InstanceURL != "" {
		lines = append(lines, fmt.Sprintf("Instance URL: %s", report.InstanceURL))
	}

	if report.AllowedUsernames == 0 {
		lines = append(lines, "Allowed usernames: unrestricted")
	} else {
		lines = append(lines, fmt.Sprintf("Allowed usernames: %d configured", report.AllowedUsernames))
	}

	if report.AdminUsernames == 0 {
		lines = append(lines, "Admin usernames: not configured")
	} else {
		lines = append(lines, fmt.Sprintf("Admin usernames: %d configured", report.AdminUsernames))
	}

	lines = append(lines, fmt.Sprintf("Linked Telegram users: %d", report.LinkedUsers))

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   strings.Join(lines, "\n"),
	})
}

func backendHealthLine(report app.HealthReport) string {
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
