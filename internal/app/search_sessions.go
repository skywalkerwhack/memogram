package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/skywalkerwhack/memogram/internal/domain"
)

func (s *Service) SearchMemosPage(ctx context.Context, telegramUserID int64, query string, offset int, limit int) (SearchPage, error) {
	if limit <= 0 {
		limit = 5
	}
	if offset < 0 {
		offset = 0
	}

	accessToken, err := s.requireAccessToken(telegramUserID)
	if err != nil {
		return SearchPage{}, err
	}

	creatorID, err := s.lookupCreatorID(ctx, accessToken)
	if err != nil {
		return SearchPage{}, err
	}

	results, err := s.backend.SearchMemos(ctx, accessToken, query, creatorID, offset+limit+1)
	if err != nil {
		return SearchPage{}, err
	}

	page := SearchPage{
		Query:  query,
		Offset: offset,
		Limit:  limit,
	}
	if offset >= len(results) {
		return page, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	page.Memos = results[offset:end]
	page.HasMore = len(results) > end
	return page, nil
}

func (s *Service) CreateSearchSession(userID int64, query string, nextOffset int, limit int) string {
	s.searchSessionMutex.Lock()
	defer s.searchSessionMutex.Unlock()

	s.searchSessionCounter++
	id := strconv.FormatUint(s.searchSessionCounter, 10)
	s.searchSessions[id] = searchSession{
		UserID: userID,
		Query:  query,
		Offset: nextOffset,
		Limit:  limit,
	}
	return id
}

func (s *Service) LoadSearchSession(sessionID string) (int64, string, int, int, bool) {
	s.searchSessionMutex.Lock()
	defer s.searchSessionMutex.Unlock()

	session, ok := s.searchSessions[sessionID]
	if !ok {
		return 0, "", 0, 0, false
	}
	return session.UserID, session.Query, session.Offset, session.Limit, true
}

func (s *Service) AdvanceSearchSession(sessionID string, nextOffset int) bool {
	s.searchSessionMutex.Lock()
	defer s.searchSessionMutex.Unlock()

	session, ok := s.searchSessions[sessionID]
	if !ok {
		return false
	}
	session.Offset = nextOffset
	s.searchSessions[sessionID] = session
	return true
}

func (s *Service) DeleteSearchSession(sessionID string) {
	s.searchSessionMutex.Lock()
	defer s.searchSessionMutex.Unlock()
	delete(s.searchSessions, sessionID)
}

func (s *Service) lookupCreatorID(ctx context.Context, accessToken string) (*int64, error) {
	user, err := s.backend.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
	}

	if user == nil {
		return nil, nil
	}
	tokens, err := domain.GetNameParentTokens(user.Name, "users/")
	if err != nil || len(tokens) != 1 {
		return nil, nil
	}

	creatorID, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return nil, nil
	}
	return &creatorID, nil
}
