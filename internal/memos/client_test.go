package memos

import (
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/skywalkerwhack/memogram/internal/domain"
	v1pb "github.com/usememos/memos/proto/gen/api/v1"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewBackendDefaultsToHTTPClient(t *testing.T) {
	backend := NewBackend("https://example.test", nil)
	if backend.httpClient != http.DefaultClient {
		t.Fatal("expected default client to be used when nil is provided")
	}
}

func TestCloneHTTPClientCopiesConfiguration(t *testing.T) {
	original := &http.Client{
		Timeout: 12 * time.Second,
	}

	cloned := cloneHTTPClient(original)
	if cloned == original {
		t.Fatal("expected a distinct client copy")
	}
	if cloned.Timeout != original.Timeout {
		t.Fatalf("expected timeout %s, got %s", original.Timeout, cloned.Timeout)
	}
}

func TestTransportOfFallsBackToDefault(t *testing.T) {
	if got := transportOf(nil); got == nil {
		t.Fatal("expected default transport for nil input")
	}
}

func TestAuthTransportAddsBearerToken(t *testing.T) {
	var gotAuth string
	transport := &authTransport{
		token: "secret-token",
		transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotAuth = req.Header.Get("Authorization")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(http.NoBody),
				Header:     make(http.Header),
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
}

func TestMemoFromProto(t *testing.T) {
	memo := memoFromProto(&v1pb.Memo{
		Name:       "memos/1",
		Content:    "content",
		Visibility: v1pb.Visibility_PROTECTED,
		Pinned:     true,
	})

	if memo.Name != "memos/1" || memo.Content != "content" || memo.Visibility != domain.VisibilityProtected || !memo.Pinned {
		t.Fatalf("unexpected memo: %#v", memo)
	}
}

func TestMemoFromProtoNil(t *testing.T) {
	memo := memoFromProto(nil)
	if memo == nil {
		t.Fatal("expected non-nil memo")
	}
}

func TestVisibilityMappings(t *testing.T) {
	if got := visibilityFromProto(v1pb.Visibility_PUBLIC); got != domain.VisibilityPublic {
		t.Fatalf("expected public, got %q", got)
	}
	if got := visibilityFromProto(v1pb.Visibility_PROTECTED); got != domain.VisibilityProtected {
		t.Fatalf("expected protected, got %q", got)
	}
	if got := visibilityToProto(domain.VisibilityPrivate); got != v1pb.Visibility_PRIVATE {
		t.Fatalf("expected private proto visibility, got %v", got)
	}
}

func TestWrapBackendErrorMapsConnectCodes(t *testing.T) {
	invalidTokenErr := wrapBackendError(connect.NewError(connect.CodeUnauthenticated, errors.New("bad token")))
	if !errors.Is(invalidTokenErr, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", invalidTokenErr)
	}

	unavailableErr := wrapBackendError(connect.NewError(connect.CodeUnavailable, errors.New("offline")))
	if !errors.Is(unavailableErr, domain.ErrBackendUnavailable) {
		t.Fatalf("expected ErrBackendUnavailable, got %v", unavailableErr)
	}
}
