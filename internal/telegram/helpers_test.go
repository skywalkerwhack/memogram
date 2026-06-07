package telegram

import (
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

func TestKeyboard(t *testing.T) {
	markup := keyboard(&domain.Memo{Name: "memos/77"})
	if len(markup.InlineKeyboard) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(markup.InlineKeyboard))
	}
	if got := markup.InlineKeyboard[0][0].CallbackData; got != "public memos/77" {
		t.Fatalf("unexpected callback data %q", got)
	}
	if got := markup.InlineKeyboard[1][0].CallbackData; got != "delete memos/77" {
		t.Fatalf("unexpected delete callback %q", got)
	}
}

func TestFormatMemoSavedMessage(t *testing.T) {
	got := formatMemoSavedMessage(domain.VisibilityPrivate, "https://example.test/", "abc")
	want := "Saved\nVisibility: *PRIVATE*\n[Open memo](https://example.test/memos/abc)"
	if got != want {
		t.Fatalf("unexpected message:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestStartUsageMessage(t *testing.T) {
	got := startUsageMessage()
	for _, want := range []string{
		"Connect Memogram to your Memos account",
		"Create an access token.",
		"/start <access_token>",
		"stored locally",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected start usage message to contain %q, got %q", want, got)
		}
	}
}

func TestTelegramMarkdownParseModeUsesMarkdownV2(t *testing.T) {
	if string(telegramMarkdownParseMode) != "MarkdownV2" {
		t.Fatalf("expected MarkdownV2 parse mode, got %q", telegramMarkdownParseMode)
	}
}

func TestReadAllLimitedRejectsOversizedReader(t *testing.T) {
	_, err := readAllLimited(strings.NewReader("abcdef"), 5)
	if err == nil {
		t.Fatal("expected oversized reader error")
	}
}

func TestReadAllLimitedAllowsReaderAtLimit(t *testing.T) {
	got, err := readAllLimited(strings.NewReader("abcde"), 5)
	if err != nil {
		t.Fatalf("readAllLimited returned error: %v", err)
	}
	if string(got) != "abcde" {
		t.Fatalf("unexpected bytes %q", got)
	}
}

func TestFormatMemoUpdatedMessage(t *testing.T) {
	got := formatMemoUpdatedMessage(domain.VisibilityProtected, "memo", "https://example.test", "abc", "📌")
	want := "Memo updated as *PROTECTED* with [memo](https://example.test/memos/abc) 📌"
	if got != want {
		t.Fatalf("unexpected message:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestFailureText(t *testing.T) {
	if got := failureText(app.ActionDelete); got != "Failed to delete memo" {
		t.Fatalf("unexpected delete failure text %q", got)
	}
	if got := failureText(app.ActionPin); got != "Failed to update memo" {
		t.Fatalf("unexpected update failure text %q", got)
	}
}

func TestBackendHealthLine(t *testing.T) {
	line := backendHealthLine(app.HealthReport{BackendAvailable: false, BackendError: "offline"})
	if line != "Backend latency: unavailable (offline)" {
		t.Fatalf("unexpected unavailable line %q", line)
	}

	line = backendHealthLine(app.HealthReport{BackendAvailable: true, BackendLatency: 1500 * time.Microsecond})
	if line != "Backend latency: 2ms" {
		t.Fatalf("unexpected latency line %q", line)
	}
}

func TestForwardedFrom(t *testing.T) {
	userOrigin := &models.MessageOrigin{
		MessageOriginUser: &models.MessageOriginUser{
			SenderUser: models.User{FirstName: "Alice", LastName: "Smith", Username: "alice"},
		},
	}
	if got := forwardedFrom(userOrigin); got.Name != "Alice Smith" || got.Username != "alice" {
		t.Fatalf("unexpected user forward info: %+v", got)
	}

	hiddenOrigin := &models.MessageOrigin{
		MessageOriginHiddenUser: &models.MessageOriginHiddenUser{SenderUserName: "Hidden"},
	}
	if got := forwardedFrom(hiddenOrigin); got.Name != "Hidden" {
		t.Fatalf("unexpected hidden forward info: %+v", got)
	}

	chatOrigin := &models.MessageOrigin{
		MessageOriginChat: &models.MessageOriginChat{SenderChat: models.Chat{Title: "Group", Username: "groupname"}},
	}
	if got := forwardedFrom(chatOrigin); got.Name != "Group" || got.Username != "groupname" {
		t.Fatalf("unexpected chat forward info: %+v", got)
	}

	channelOrigin := &models.MessageOrigin{
		MessageOriginChannel: &models.MessageOriginChannel{Chat: models.Chat{Title: "Channel", Username: "channelname"}},
	}
	if got := forwardedFrom(channelOrigin); got.Name != "Channel" || got.Username != "channelname" {
		t.Fatalf("unexpected channel forward info: %+v", got)
	}
}
