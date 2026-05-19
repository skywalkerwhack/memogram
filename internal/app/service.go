package app

import (
	"context"
	"strings"
	"sync"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

type Service struct {
	backend Backend
	store   TokenStore

	serverURL        string
	dataFile         string
	allowedUsernames map[string]struct{}

	mediaGroupCache sync.Map
	mediaGroupMutex sync.Mutex

	instanceProfile *domain.InstanceProfile
}

func NewService(backend Backend, store TokenStore, dataFile string, allowedUsernames []string) *Service {
	allowedSet := make(map[string]struct{}, len(allowedUsernames))
	for _, username := range allowedUsernames {
		allowedSet[strings.ToLower(strings.TrimSpace(username))] = struct{}{}
	}

	return &Service{
		backend:          backend,
		store:            store,
		serverURL:        backend.BaseURL(),
		dataFile:         dataFile,
		allowedUsernames: allowedSet,
	}
}

func (s *Service) Start(ctx context.Context) {
	profile, err := s.backend.GetInstanceProfile(ctx)
	if err == nil {
		s.instanceProfile = profile
	}
}

func (s *Service) IsUserAllowed(username string) bool {
	if len(s.allowedUsernames) == 0 {
		return true
	}
	if username == "" {
		return false
	}
	_, ok := s.allowedUsernames[strings.ToLower(strings.TrimSpace(username))]
	return ok
}

func (s *Service) MemoBaseURL() string {
	if s.instanceProfile != nil && s.instanceProfile.InstanceURL != "" {
		return s.instanceProfile.InstanceURL
	}
	return s.serverURL
}

func (s *Service) requireAccessToken(telegramUserID int64) (string, error) {
	accessToken, ok := s.store.GetUserAccessToken(telegramUserID)
	if !ok {
		return "", domain.ErrAccountNotLinked
	}
	return accessToken, nil
}
