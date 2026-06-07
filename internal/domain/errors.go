package domain

import "errors"

var (
	ErrAccountNotLinked   = errors.New("account not linked")
	ErrBackendUnavailable = errors.New("backend unavailable")
	ErrInvalidToken       = errors.New("invalid access token")
	ErrUserNotAllowed     = errors.New("user not allowed")
)
