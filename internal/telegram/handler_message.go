package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/domain"
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

	if updated := t.tryHandlePendingEdit(ctx, message, content, hasAttachment); updated {
		return
	}

	if content == "" && !hasAttachment {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Please input memo content",
		})
		return
	}

	input := app.CreateMemoInput{
		UserID:  message.From.ID,
		Content: content,
	}
	if message.MediaGroupID != "" {
		input.AttachmentSet = message.MediaGroupID
	}
	if message.ForwardOrigin != nil {
		input.ForwardedFrom = forwardedFrom(message.ForwardOrigin)
	}

	memo, err := t.service.CreateMemo(ctx, input)
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

	t.sendMemoCard(ctx, message.Chat.ID, *memo, message.ID)
}

func formatMemoSavedMessage(visibility domain.Visibility, memoName string, baseURL string, memoUID string) string {
	return fmt.Sprintf(
		"Content saved as *%s* with [%s](%s)",
		escapeMarkdownV2(string(visibility)),
		escapeMarkdownV2(memoName),
		escapeMarkdownV2URL(strings.TrimRight(baseURL, "/")+"/memos/"+memoUID),
	)
}

func (t *Bot) tryHandlePendingEdit(ctx context.Context, message *models.Message, content string, hasAttachment bool) bool {
	if !t.service.HasPendingMemoEdit(message.From.ID) {
		return false
	}

	if hasAttachment {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Editing only accepts text. Send the replacement text or /cancel.",
		})
		return true
	}
	if strings.TrimSpace(content) == "" {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Send the replacement text for the memo, or /cancel.",
		})
		return true
	}

	updatedMemo, err := t.service.UpdatePendingMemoContent(ctx, message.From.ID, content)
	if err == nil {
		t.sendMemoUpdateCard(ctx, message.Chat.ID, *updatedMemo, message.ID)
		return true
	}
	if errors.Is(err, domain.ErrMemoEditNotFound) {
		return false
	}
	if errors.Is(err, domain.ErrAccountNotLinked) {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Please start the bot with /start <access_token>",
		})
		return true
	}

	t.sendError(message.Chat.ID, fmt.Errorf("failed to update memo: %w", err))
	return true
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
