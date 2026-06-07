package memos

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/skywalkerwhack/memogram/internal/domain"
	v1pb "github.com/usememos/memos/proto/gen/api/v1"
	"github.com/usememos/memos/proto/gen/api/v1/apiv1connect"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
)

type Backend struct {
	baseURL    string
	httpClient *http.Client
}

func NewBackend(baseURL string, httpClient *http.Client) *Backend {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Backend{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (b *Backend) BaseURL() string {
	return b.baseURL
}

func (b *Backend) GetInstanceProfile(ctx context.Context) (*domain.InstanceProfile, error) {
	resp, err := b.newClient("").InstanceService.GetInstanceProfile(ctx, connect.NewRequest(&v1pb.GetInstanceProfileRequest{}))
	if err != nil {
		return nil, wrapBackendError(err)
	}
	return &domain.InstanceProfile{InstanceURL: resp.Msg.GetInstanceUrl()}, nil
}

func (b *Backend) GetCurrentUser(ctx context.Context, accessToken string) (*domain.User, error) {
	resp, err := b.newClient(accessToken).AuthService.GetCurrentUser(ctx, connect.NewRequest(&v1pb.GetCurrentUserRequest{}))
	if err != nil {
		return nil, wrapBackendError(err)
	}
	user := resp.Msg.GetUser()
	if user == nil {
		return &domain.User{}, nil
	}
	return &domain.User{
		Name:        user.GetName(),
		Username:    user.GetUsername(),
		DisplayName: user.GetDisplayName(),
	}, nil
}

func (b *Backend) CreateMemo(ctx context.Context, accessToken string, content string) (*domain.Memo, error) {
	resp, err := b.newClient(accessToken).MemoService.CreateMemo(ctx, connect.NewRequest(&v1pb.CreateMemoRequest{
		Memo: &v1pb.Memo{Content: content},
	}))
	if err != nil {
		return nil, wrapBackendError(err)
	}
	return memoFromProto(resp.Msg), nil
}

func (b *Backend) GetMemo(ctx context.Context, accessToken string, name string) (*domain.Memo, error) {
	resp, err := b.newClient(accessToken).MemoService.GetMemo(ctx, connect.NewRequest(&v1pb.GetMemoRequest{Name: name}))
	if err != nil {
		return nil, wrapBackendError(err)
	}
	return memoFromProto(resp.Msg), nil
}

func (b *Backend) UpdateMemo(ctx context.Context, accessToken string, memo *domain.Memo) (*domain.Memo, error) {
	resp, err := b.newClient(accessToken).MemoService.UpdateMemo(ctx, connect.NewRequest(&v1pb.UpdateMemoRequest{
		Memo: &v1pb.Memo{
			Name:       memo.Name,
			Content:    memo.Content,
			Visibility: visibilityToProto(memo.Visibility),
			Pinned:     memo.Pinned,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"visibility", "pinned"}},
	}))
	if err != nil {
		return nil, wrapBackendError(err)
	}
	return memoFromProto(resp.Msg), nil
}

func (b *Backend) DeleteMemo(ctx context.Context, accessToken string, name string) error {
	_, err := b.newClient(accessToken).MemoService.DeleteMemo(ctx, connect.NewRequest(&v1pb.DeleteMemoRequest{
		Name:  name,
		Force: true,
	}))
	return wrapBackendError(err)
}

func (b *Backend) SearchMemos(ctx context.Context, accessToken string, query string, creatorID *int64, limit int) ([]domain.Memo, error) {
	filter := fmt.Sprintf("content.contains(%s)", strconv.Quote(query))
	if creatorID != nil {
		filter = fmt.Sprintf("%s && creator_id == %d", filter, *creatorID)
	}

	resp, err := b.newClient(accessToken).MemoService.ListMemos(ctx, connect.NewRequest(&v1pb.ListMemosRequest{
		PageSize: int32(limit),
		Filter:   filter,
	}))
	if err != nil {
		return nil, wrapBackendError(err)
	}

	memos := make([]domain.Memo, 0, len(resp.Msg.GetMemos()))
	for _, memo := range resp.Msg.GetMemos() {
		memos = append(memos, *memoFromProto(memo))
	}
	return memos, nil
}

func (b *Backend) UploadAttachment(ctx context.Context, accessToken string, memoName string, file domain.FilePayload) error {
	_, err := b.newClient(accessToken).AttachmentService.CreateAttachment(ctx, connect.NewRequest(&v1pb.CreateAttachmentRequest{
		Attachment: &v1pb.Attachment{
			Filename: file.Filename,
			Type:     file.ContentType,
			Size:     int64(len(file.Bytes)),
			Content:  file.Bytes,
			Memo:     &memoName,
		},
	}))
	return wrapBackendError(err)
}

func wrapBackendError(err error) error {
	if err == nil {
		return nil
	}

	switch connect.CodeOf(err) {
	case connect.CodeUnauthenticated:
		return fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
	case connect.CodeDeadlineExceeded, connect.CodeUnavailable:
		return fmt.Errorf("%w: %v", domain.ErrBackendUnavailable, err)
	default:
		return err
	}
}

type client struct {
	InstanceService   apiv1connect.InstanceServiceClient
	AuthService       apiv1connect.AuthServiceClient
	MemoService       apiv1connect.MemoServiceClient
	AttachmentService apiv1connect.AttachmentServiceClient
}

func (b *Backend) newClient(accessToken string) *client {
	httpClient := b.httpClient
	if accessToken != "" {
		httpClient = cloneHTTPClient(httpClient)
		httpClient.Transport = &authTransport{
			token:     accessToken,
			transport: transportOf(httpClient.Transport),
		}
	}

	return &client{
		InstanceService:   apiv1connect.NewInstanceServiceClient(httpClient, b.baseURL),
		AuthService:       apiv1connect.NewAuthServiceClient(httpClient, b.baseURL),
		MemoService:       apiv1connect.NewMemoServiceClient(httpClient, b.baseURL),
		AttachmentService: apiv1connect.NewAttachmentServiceClient(httpClient, b.baseURL),
	}
}

func cloneHTTPClient(httpClient *http.Client) *http.Client {
	if httpClient == nil {
		return http.DefaultClient
	}

	cloned := *httpClient
	return &cloned
}

func transportOf(transport http.RoundTripper) http.RoundTripper {
	if transport != nil {
		return transport
	}
	return http.DefaultTransport
}

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

func memoFromProto(memo *v1pb.Memo) *domain.Memo {
	if memo == nil {
		return &domain.Memo{}
	}
	return &domain.Memo{
		Name:       memo.GetName(),
		Content:    memo.GetContent(),
		Visibility: visibilityFromProto(memo.GetVisibility()),
		Pinned:     memo.GetPinned(),
	}
}

func visibilityFromProto(visibility v1pb.Visibility) domain.Visibility {
	switch visibility {
	case v1pb.Visibility_PUBLIC:
		return domain.VisibilityPublic
	case v1pb.Visibility_PROTECTED:
		return domain.VisibilityProtected
	default:
		return domain.VisibilityPrivate
	}
}

func visibilityToProto(visibility domain.Visibility) v1pb.Visibility {
	switch visibility {
	case domain.VisibilityPublic:
		return v1pb.Visibility_PUBLIC
	case domain.VisibilityProtected:
		return v1pb.Visibility_PROTECTED
	default:
		return v1pb.Visibility_PRIVATE
	}
}
