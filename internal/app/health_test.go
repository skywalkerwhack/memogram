package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

type testBackend struct {
	getInstanceProfile func(context.Context) (*domain.InstanceProfile, error)
}

func (b testBackend) BaseURL() string { return "https://example.test" }

func (b testBackend) GetInstanceProfile(ctx context.Context) (*domain.InstanceProfile, error) {
	return b.getInstanceProfile(ctx)
}

func (b testBackend) GetCurrentUser(context.Context, string) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (b testBackend) CreateMemo(context.Context, string, string) (*domain.Memo, error) {
	return nil, errors.New("not implemented")
}

func (b testBackend) GetMemo(context.Context, string, string) (*domain.Memo, error) {
	return nil, errors.New("not implemented")
}

func (b testBackend) UpdateMemo(context.Context, string, *domain.Memo) (*domain.Memo, error) {
	return nil, errors.New("not implemented")
}

func (b testBackend) DeleteMemo(context.Context, string, string) error {
	return errors.New("not implemented")
}

func (b testBackend) SearchMemos(context.Context, string, string, *int64, int) ([]domain.Memo, error) {
	return nil, errors.New("not implemented")
}

func (b testBackend) UploadAttachment(context.Context, string, string, domain.FilePayload) error {
	return errors.New("not implemented")
}

func TestProbeBackendLatencySuccess(t *testing.T) {
	backend := testBackend{
		getInstanceProfile: func(context.Context) (*domain.InstanceProfile, error) {
			time.Sleep(15 * time.Millisecond)
			return &domain.InstanceProfile{InstanceURL: "https://example.test"}, nil
		},
	}

	status := ProbeBackendLatency(context.Background(), backend)
	if status.Err != nil {
		t.Fatalf("expected no error, got %v", status.Err)
	}
	if status.Latency < 15*time.Millisecond {
		t.Fatalf("expected latency >= 15ms, got %s", status.Latency)
	}

	line := status.StatusLine()
	if !strings.HasPrefix(line, "Backend latency: ") {
		t.Fatalf("unexpected status line: %q", line)
	}
	if strings.Contains(line, "unavailable") {
		t.Fatalf("expected reachable backend status line, got %q", line)
	}
}

func TestProbeBackendLatencyFailure(t *testing.T) {
	backend := testBackend{
		getInstanceProfile: func(context.Context) (*domain.InstanceProfile, error) {
			return nil, errors.New("backend offline")
		},
	}

	status := ProbeBackendLatency(context.Background(), backend)
	if status.Err == nil {
		t.Fatal("expected an error")
	}

	line := status.StatusLine()
	if !strings.Contains(line, "unavailable") {
		t.Fatalf("expected unavailable status line, got %q", line)
	}
	if !strings.Contains(line, "backend offline") {
		t.Fatalf("expected error message in status line, got %q", line)
	}
}

func TestProbeBackendLatencyWithNilClient(t *testing.T) {
	status := ProbeBackendLatency(context.Background(), nil)
	if status.Err == nil {
		t.Fatal("expected an error for nil client")
	}
	if got := status.StatusLine(); !strings.Contains(got, "unavailable") {
		t.Fatalf("expected unavailable status, got %q", got)
	}
}
