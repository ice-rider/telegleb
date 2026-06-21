package auth

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidPhone         = errors.New("invalid phone number format")
	ErrAuthFailed           = errors.New("failed to initiate authentication")
	ErrSessionNotFound      = errors.New("auth session not found")
	ErrInvalidStep          = errors.New("invalid authentication step")
	ErrInvalidSessionState  = errors.New("invalid session state for operation")
)