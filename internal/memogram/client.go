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
	"net/http"

	"github.com/usememos/memos/proto/gen/api/v1/apiv1connect"
)

type MemosClient struct {
	baseURL string

	InstanceService   apiv1connect.InstanceServiceClient
	AuthService       apiv1connect.AuthServiceClient
	UserService       apiv1connect.UserServiceClient
	MemoService       apiv1connect.MemoServiceClient
	AttachmentService apiv1connect.AttachmentServiceClient
}

// NewMemosClient creates a new client using Connect protocol
// baseURL should be the full HTTP URL (e.g., "http://localhost:8081")
func NewMemosClient(baseURL string) *MemosClient {
	httpClient := http.DefaultClient

	return &MemosClient{
		baseURL:           baseURL,
		InstanceService:   apiv1connect.NewInstanceServiceClient(httpClient, baseURL),
		AuthService:       apiv1connect.NewAuthServiceClient(httpClient, baseURL),
		UserService:       apiv1connect.NewUserServiceClient(httpClient, baseURL),
		MemoService:       apiv1connect.NewMemoServiceClient(httpClient, baseURL),
		AttachmentService: apiv1connect.NewAttachmentServiceClient(httpClient, baseURL),
	}
}

// NewAuthenticatedClient creates a new client with authentication
func (c *MemosClient) NewAuthenticatedClient(accessToken string) *MemosClient {
	httpClient := &http.Client{
		Transport: &authTransport{
			token:     accessToken,
			transport: http.DefaultTransport,
		},
	}

	return &MemosClient{
		baseURL:           c.baseURL,
		InstanceService:   apiv1connect.NewInstanceServiceClient(httpClient, c.baseURL),
		AuthService:       apiv1connect.NewAuthServiceClient(httpClient, c.baseURL),
		UserService:       apiv1connect.NewUserServiceClient(httpClient, c.baseURL),
		MemoService:       apiv1connect.NewMemoServiceClient(httpClient, c.baseURL),
		AttachmentService: apiv1connect.NewAttachmentServiceClient(httpClient, c.baseURL),
	}
}

// authTransport adds Authorization header to all HTTP requests
type authTransport struct {
	token     string
	transport http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.transport.RoundTrip(req)
}
