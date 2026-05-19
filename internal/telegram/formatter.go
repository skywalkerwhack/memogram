package telegram

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/go-telegram/bot/models"
)

func formatContent(content string, contentEntities []models.MessageEntity) string {
	sort.Slice(contentEntities, func(i, j int) bool {
		if contentEntities[i].Offset == contentEntities[j].Offset {
			return contentEntities[i].Length < contentEntities[j].Length
		}
		return contentEntities[i].Offset < contentEntities[j].Offset
	})

	contentRunes := utf16.Encode([]rune(content))

	var sb strings.Builder
	cursor := 0
	for _, entity := range contentEntities {
		if !isSupportedEntity(entity.Type) {
			continue
		}
		start := entity.Offset
		end := entity.Offset + entity.Length
		if start < cursor {
			continue
		}
		if start >= len(contentRunes) {
			break
		}
		if end > len(contentRunes) {
			end = len(contentRunes)
		}

		sb.WriteString(escapeMarkdownV2(string(utf16.Decode(contentRunes[cursor:start]))))
		segment := string(utf16.Decode(contentRunes[start:end]))
		sb.WriteString(applyEntityFormatting(segment, entity))
		cursor = end
	}
	sb.WriteString(escapeMarkdownV2(string(utf16.Decode(contentRunes[cursor:]))))
	return sb.String()
}

func isSupportedEntity(entityType models.MessageEntityType) bool {
	switch entityType {
	case models.MessageEntityTypeURL,
		models.MessageEntityTypeTextLink,
		models.MessageEntityTypeBold,
		models.MessageEntityTypeItalic:
		return true
	default:
		return false
	}
}

func applyEntityFormatting(segment string, entity models.MessageEntity) string {
	if strings.TrimSpace(segment) == "" {
		return segment
	}
	prefix, core, suffix := trimMarkdownSegment(segment)
	switch entity.Type {
	case models.MessageEntityTypeURL:
		return fmt.Sprintf("%s[%s](%s)%s", escapeMarkdownV2(prefix), escapeMarkdownV2(core), escapeMarkdownV2URL(core), escapeMarkdownV2(suffix))
	case models.MessageEntityTypeTextLink:
		return fmt.Sprintf("%s[%s](%s)%s", escapeMarkdownV2(prefix), escapeMarkdownV2(core), escapeMarkdownV2URL(entity.URL), escapeMarkdownV2(suffix))
	case models.MessageEntityTypeBold:
		return fmt.Sprintf("%s*%s*%s", escapeMarkdownV2(prefix), escapeMarkdownV2(core), escapeMarkdownV2(suffix))
	case models.MessageEntityTypeItalic:
		return fmt.Sprintf("%s_%s_%s", escapeMarkdownV2(prefix), escapeMarkdownV2(core), escapeMarkdownV2(suffix))
	default:
		return escapeMarkdownV2(segment)
	}
}

func trimMarkdownSegment(segment string) (string, string, string) {
	start := 0
	for start < len(segment) {
		if segment[start] != ' ' && segment[start] != '\n' && segment[start] != '\t' {
			break
		}
		start++
	}

	end := len(segment)
	for end > start {
		if segment[end-1] != ' ' && segment[end-1] != '\n' && segment[end-1] != '\t' {
			break
		}
		end--
	}

	return segment[:start], segment[start:end], segment[end:]
}

func escapeMarkdownV2(text string) string {
	var replacer = strings.NewReplacer(
		"\\", "\\\\",
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func escapeMarkdownV2URL(text string) string {
	var replacer = strings.NewReplacer(
		"\\", "\\\\",
		")", "\\)",
	)
	return replacer.Replace(text)
}
