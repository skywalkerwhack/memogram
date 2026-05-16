package domain

import "errors"

var (
	ErrAccountNotLinked = errors.New("account not linked")
	ErrInvalidToken     = errors.New("invalid access token")
	ErrUserNotAllowed   = errors.New("user not allowed")
)
