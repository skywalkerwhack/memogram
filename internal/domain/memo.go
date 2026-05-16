package domain

import (
	"fmt"
	"strings"
)

type Visibility string

const (
	VisibilityPublic    Visibility = "PUBLIC"
	VisibilityProtected Visibility = "PROTECTED"
	VisibilityPrivate   Visibility = "PRIVATE"
)

type Memo struct {
	Name       string
	Content    string
	Visibility Visibility
	Pinned     bool
}

type User struct {
	Name        string
	Username    string
	DisplayName string
}

type InstanceProfile struct {
	InstanceURL string
}

type FilePayload struct {
	Filename    string
	ContentType string
	Bytes       []byte
}

type ForwardInfo struct {
	Name     string
	Username string
}

func GetNameParentTokens(name string, tokenPrefixes ...string) ([]string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2*len(tokenPrefixes) {
		return nil, fmt.Errorf("invalid request %q", name)
	}

	var tokens []string
	for i, tokenPrefix := range tokenPrefixes {
		if fmt.Sprintf("%s/", parts[2*i]) != tokenPrefix {
			return nil, fmt.Errorf("invalid prefix %q in request %q", tokenPrefix, name)
		}
		if parts[2*i+1] == "" {
			return nil, fmt.Errorf("invalid request %q with empty prefix %q", name, tokenPrefix)
		}
		tokens = append(tokens, parts[2*i+1])
	}
	return tokens, nil
}

func ExtractMemoUIDFromName(name string) (string, error) {
	tokens, err := GetNameParentTokens(name, "memos/")
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}
