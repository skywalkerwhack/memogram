package telegram

import (
	"testing"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/skywalkerwhack/memogram/internal/app"
	"github.com/skywalkerwhack/memogram/internal/domain"
)

func TestKeyboard(t *testing.T) {
	markup := keyboard(&domain.Memo{Name: "memos/77"})
	if len(markup.InlineKeyboard) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(markup.InlineKeyboard))
	}
	if got := markup.InlineKeyboard[0][0].CallbackData; got != "public memos/77" {
		t.Fatalf("unexpected callback data %q", got)
	}
	if got := markup.InlineKeyboard[0][1].CallbackData; got != "protected memos/77" {
		t.Fatalf("unexpected protected callback %q", got)
	}
	if got := markup.InlineKeyboard[1][1].CallbackData; got != "edit memos/77" {
		t.Fatalf("unexpected edit callback %q", got)
	}
	if got := markup.InlineKeyboard[2][0].CallbackData; got != "delete memos/77" {
		t.Fatalf("unexpected delete callback %q", got)
	}
}

func TestFormatMemoCard(t *testing.T) {
	got := formatMemoCard(domain.Memo{
		Name:       "memo_[1]",
		Content:    "line one",
		Visibility: domain.VisibilityPrivate,
		Pinned:     true,
	}, "https://example.test/")
	want := "*PRIVATE* [memo\\_\\[1\\]](https://example.test) 📌\nline one"
	if got != want {
		t.Fatalf("unexpected message:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestFormatMemoUpdatedCard(t *testing.T) {
	got := formatMemoUpdatedCard(domain.Memo{
		Name:       "memos/abc",
		Content:    "content",
		Visibility: domain.VisibilityProtected,
		Pinned:     true,
	}, "https://example.test")
	want := "Memo updated\n*PROTECTED* [memos/abc](https://example.test/memos/abc) 📌\ncontent"
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

func TestSplitCallbackData(t *testing.T) {
	parts := splitCallbackData("searchmore 12")
	if len(parts) != 2 || parts[0] != "searchmore" || parts[1] != "12" {
		t.Fatalf("unexpected split result: %#v", parts)
	}
}
