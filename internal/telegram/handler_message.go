package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/usememos/memogram/internal/domain"
)

func (t *Bot) handleMessage(ctx context.Context, update *models.Update) {
	message := update.Message
	hasAttachment := message.Document != nil || len(message.Photo) > 0 || message.Voice != nil || message.Video != nil

	content := message.Text
	contentEntities := message.Entities
	if message.Caption != "" {
		content = message.Caption
		contentEntities = message.CaptionEntities
	}
	if len(contentEntities) > 0 {
		content = formatContent(content, contentEntities)
	}

	if content == "" && !hasAttachment {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Please input memo content",
		})
		return
	}

	input := domain.SaveMessageInput{
		TelegramUserID: message.From.ID,
		Content:        content,
	}
	if message.MediaGroupID != "" {
		input.MediaGroupID = message.MediaGroupID
	}
	if message.ForwardOrigin != nil {
		input.ForwardedFrom = forwardedFrom(message.ForwardOrigin)
	}

	memo, err := t.service.CreateMemoFromMessage(ctx, input)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotLinked) {
			t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   "Please start the bot with /start <access_token>",
			})
			return
		}
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Failed to create memo",
		})
		return
	}

	fileIDs := make([]string, 0, 4)
	if message.Document != nil {
		fileIDs = append(fileIDs, message.Document.FileID)
	}
	if message.Voice != nil {
		fileIDs = append(fileIDs, message.Voice.FileID)
	}
	if message.Video != nil {
		fileIDs = append(fileIDs, message.Video.FileID)
	}
	if len(message.Photo) > 0 {
		fileIDs = append(fileIDs, message.Photo[len(message.Photo)-1].FileID)
	}

	for _, fileID := range fileIDs {
		payload, err := t.fetchFilePayload(ctx, fileID)
		if err != nil {
			t.sendError(message.Chat.ID, fmt.Errorf("failed to get file: %w", err))
			return
		}
		if err := t.service.AttachFile(ctx, message.From.ID, memo.Name, payload); err != nil {
			t.sendError(message.Chat.ID, fmt.Errorf("failed to save attachment: %w", err))
			return
		}
	}

	memoUID, err := domain.ExtractMemoUIDFromName(memo.Name)
	if err != nil {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Failed to save memo",
		})
		return
	}

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:              message.Chat.ID,
		Text:                fmt.Sprintf("Content saved as %s with [%s](%s/memos/%s)", memo.Visibility, memo.Name, t.service.MemoBaseURL(), memoUID),
		ParseMode:           models.ParseModeMarkdown,
		DisableNotification: true,
		ReplyParameters: &models.ReplyParameters{
			MessageID: message.ID,
		},
		ReplyMarkup: keyboard(memo),
	})
}

func (t *Bot) fetchFilePayload(ctx context.Context, fileID string) (domain.FilePayload, error) {
	file, err := t.bot.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return domain.FilePayload{}, err
	}

	fileLink := t.bot.FileDownloadLink(file)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileLink, nil)
	if err != nil {
		return domain.FilePayload{}, fmt.Errorf("create download request: %w", err)
	}

	response, err := t.httpClient.Do(req)
	if err != nil {
		return domain.FilePayload{}, fmt.Errorf("download file: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		return domain.FilePayload{}, fmt.Errorf("download failed with status %s", response.Status)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return domain.FilePayload{}, fmt.Errorf("read file: %w", err)
	}

	contentType := response.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(bytes)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return domain.FilePayload{
		Filename:    filepath.Base(file.FilePath),
		ContentType: contentType,
		Bytes:       bytes,
	}, nil
}

func forwardedFrom(origin *models.MessageOrigin) *domain.ForwardInfo {
	info := &domain.ForwardInfo{}

	switch {
	case origin.MessageOriginUser != nil:
		user := origin.MessageOriginUser.SenderUser
		if user.LastName != "" {
			info.Name = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		} else {
			info.Name = user.FirstName
		}
		info.Username = user.Username
	case origin.MessageOriginHiddenUser != nil:
		if origin.MessageOriginHiddenUser.SenderUserName != "" {
			info.Name = origin.MessageOriginHiddenUser.SenderUserName
		} else {
			info.Name = "Hidden User"
		}
	case origin.MessageOriginChat != nil:
		info.Name = origin.MessageOriginChat.SenderChat.Title
		info.Username = origin.MessageOriginChat.SenderChat.Username
	case origin.MessageOriginChannel != nil:
		info.Name = origin.MessageOriginChannel.Chat.Title
		info.Username = origin.MessageOriginChannel.Chat.Username
	}

	return info
}
