package auth

import "errors"

var (
	ErrInvalidPhone        = errors.New("phone number must start with + and contain digits")
	ErrInvalidCode         = errors.New("invalid confirmation code")
	ErrInvalidPassword     = errors.New("invalid 2FA password")
	ErrEmptyInput          = errors.New("required field is empty")
	ErrSessionNotFound     = errors.New("session not found or expired")
	ErrInvalidSessionState = errors.New("authorization step called out of order")
)
