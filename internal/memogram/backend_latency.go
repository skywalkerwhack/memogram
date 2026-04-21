// memogram is a Telegram bot for saving messages into a Memos instance.
// Copyright (C) 2026  skywalkerwhack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package memogram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	v1pb "github.com/usememos/memos/proto/gen/api/v1"
)

type BackendLatencyStatus struct {
	Latency time.Duration
	Err     error
}

func ProbeBackendLatency(ctx context.Context, client *MemosClient) BackendLatencyStatus {
	if client == nil || client.InstanceService == nil {
		return BackendLatencyStatus{Err: fmt.Errorf("instance service is not configured")}
	}

	start := time.Now()
	_, err := client.InstanceService.GetInstanceProfile(ctx, connect.NewRequest(&v1pb.GetInstanceProfileRequest{}))
	return BackendLatencyStatus{
		Latency: time.Since(start),
		Err:     err,
	}
}

func (s BackendLatencyStatus) StatusLine() string {
	if s.Err != nil {
		return fmt.Sprintf("Backend latency: unavailable (%s)", sanitizeBackendError(s.Err))
	}
	return fmt.Sprintf("Backend latency: %s", formatLatency(s.Latency))
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

func sanitizeBackendError(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "unknown error"
	}
	return message
}
