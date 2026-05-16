package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/usememos/memogram/internal/app"
	"github.com/usememos/memogram/internal/config"
)

type Bot struct {
	bot        *bot.Bot
	service    *app.Service
	config     *config.Config
	httpClient *http.Client
}

func NewBot(cfg *config.Config, service *app.Service) (*Bot, error) {
	tg := &Bot{
		service:    service,
		config:     cfg,
		httpClient: http.DefaultClient,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(tg.handleUpdate),
		bot.WithCallbackQueryDataHandler("", bot.MatchTypePrefix, tg.handleCallbackQuery),
	}
	if cfg.BotProxyAddr != "" {
		opts = append(opts, bot.WithServerURL(cfg.BotProxyAddr))
	}

	b, err := bot.New(cfg.BotToken, opts...)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}
	tg.bot = b

	return tg, nil
}

func (t *Bot) Start(ctx context.Context) {
	t.service.Start(ctx)
	slog.Info("Memogram started")

	commands := []models.BotCommand{
		{Command: "start", Description: "Start the bot with access token"},
		{Command: "help", Description: "Show available commands"},
		{Command: "unlink", Description: "Disconnect your linked account"},
		{Command: "search", Description: "Search for the memos"},
		{Command: "status", Description: "Show bot and account status"},
		{Command: "ping", Description: "Ping the bot"},
	}
	if _, err := t.bot.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: commands}); err != nil {
		slog.Error("failed to set bot commands", slog.Any("err", err))
	}

	t.bot.Start(ctx)
}

func (t *Bot) handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil || update.Message.From == nil {
		t.sendError(0, errors.New("invalid message structure: missing required fields"))
		return
	}
	if update.Message.Chat.ID == 0 {
		t.sendError(0, errors.New("invalid chat: missing chat ID"))
		return
	}

	username := update.Message.From.Username
	if !t.service.IsUserAllowed(username) {
		if username == "" {
			t.sendError(update.Message.Chat.ID, errors.New("your account must have a username to use this bot"))
			return
		}
		t.sendError(update.Message.Chat.ID, fmt.Errorf("your account %s is not allowed to use this bot", username))
		return
	}

	if t.handleCommand(ctx, update) {
		return
	}
	t.handleMessage(ctx, update)
}

func (t *Bot) sendError(chatID int64, err error) {
	slog.Error("error", slog.Any("err", err))
	if chatID == 0 {
		return
	}
	t.bot.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("Error: %s", err.Error()),
	})
}
