package telegram

import "errors"

const (
	ErrCodeSessionPasswordNeeded = "SESSION_PASSWORD_NEEDED"
)

var (
	ErrClientInitFailed     = errors.New("telegram client initialization failed")
	ErrTelegramSendCode     = errors.New("telegram api error: failed to send verification code")
	ErrUnexpectedCodeType   = errors.New("telegram api error: unexpected response type on send code")
	ErrTelegramSignIn       = errors.New("telegram api error: failed to sign in with code")
	ErrUnexpectedSignInType = errors.New("telegram api error: unexpected response type on sign in")
	ErrInvalid2FAPassword   = errors.New("telegram api error: invalid 2FA password")
	ErrSessionNotFound      = errors.New("session not found for the given token")
)
