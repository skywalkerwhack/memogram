package store

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type FileTokenStore struct {
	dataPath string

	userAccessTokenCache sync.Map // map[int64]string
}

func NewFileTokenStore(dataPath string) (*FileTokenStore, error) {
	store := &FileTokenStore{dataPath: dataPath}
	if err := store.loadUserAccessTokenMapFromFile(); err != nil {
		return nil, fmt.Errorf("load user access token map from file: %w", err)
	}
	return store, nil
}

func (s *FileTokenStore) GetUserAccessToken(userID int64) (string, bool) {
	accessToken, ok := s.userAccessTokenCache.Load(userID)
	if !ok {
		return "", false
	}
	return accessToken.(string), true
}

func (s *FileTokenStore) SetUserAccessToken(userID int64, accessToken string) error {
	s.userAccessTokenCache.Store(userID, accessToken)
	return s.saveUserAccessTokenMapToFile()
}

func (s *FileTokenStore) DeleteUserAccessToken(userID int64) (bool, error) {
	_, existed := s.userAccessTokenCache.LoadAndDelete(userID)
	if !existed {
		return false, nil
	}
	if err := s.saveUserAccessTokenMapToFile(); err != nil {
		return true, err
	}
	return true, nil
}

func (s *FileTokenStore) CountUserAccessTokens() int {
	return len(s.snapshotAccessTokens())
}

func (s *FileTokenStore) saveUserAccessTokenMapToFile() error {
	entries := s.snapshotAccessTokens()
	dataDir := filepath.Dir(s.dataPath)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	tmpFile, err := os.CreateTemp(dataDir, "memogram-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	writer := bufio.NewWriter(tmpFile)
	for _, entry := range entries {
		if _, err := fmt.Fprintf(writer, "%d:%s\n", entry.userID, entry.accessToken); err != nil {
			tmpFile.Close()
			return fmt.Errorf("write data file: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("flush data file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("sync data file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close data file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), s.dataPath); err != nil {
		return fmt.Errorf("replace data file: %w", err)
	}
	return nil
}

func (s *FileTokenStore) loadUserAccessTokenMapFromFile() error {
	if err := os.MkdirAll(filepath.Dir(s.dataPath), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(s.dataPath); os.IsNotExist(err) {
		file, err := os.OpenFile(s.dataPath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		file.Close()
	} else if err != nil {
		return err
	}

	file, err := os.Open(s.dataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		userID, accessToken := parseLine(line)
		if userID == 0 || accessToken == "" {
			continue
		}
		s.userAccessTokenCache.Store(userID, accessToken)
	}
	return scanner.Err()
}

func parseLine(line string) (int64, string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return 0, ""
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, ""
	}
	return userID, parts[1]
}

type userAccessTokenEntry struct {
	userID      int64
	accessToken string
}

func (s *FileTokenStore) snapshotAccessTokens() []userAccessTokenEntry {
	entries := make([]userAccessTokenEntry, 0)
	s.userAccessTokenCache.Range(func(key, value any) bool {
		userID, ok := key.(int64)
		if !ok {
			return true
		}
		accessToken, ok := value.(string)
		if !ok {
			return true
		}
		entries = append(entries, userAccessTokenEntry{
			userID:      userID,
			accessToken: accessToken,
		})
		return true
	})

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].userID < entries[j].userID
	})

	return entries
}
