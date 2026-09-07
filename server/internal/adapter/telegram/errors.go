package telegram

import (
	"errors"
	"time"

	"github.com/gotd/td/tgerr"
)

const ErrCodeSessionPasswordNeeded = "SESSION_PASSWORD_NEEDED"

var (
	ErrClientInitFailed     = errors.New("telegram client initialization failed")
	ErrTelegramSendCode     = errors.New("telegram api error: failed to send verification code")
	ErrUnexpectedCodeType   = errors.New("telegram api error: unexpected response type on send code")
	ErrTelegramSignIn       = errors.New("telegram api error: failed to sign in with code")
	ErrUnexpectedSignInType = errors.New("telegram api error: unexpected response type on sign in")
	ErrInvalid2FAPassword   = errors.New("telegram api error: invalid 2FA password")
	ErrSessionNotFound      = errors.New("session not found for the given token")
)

// FloodWait — превышение лимита запросов на стороне Telegram. Оно требует
// отдельной обработки: клиенту нужно сказать, через сколько повторять, а не
// отдавать безликую пятисотку.
type FloodWait struct {
	RetryAfter int
}

func (e *FloodWait) Error() string {
	return "telegram rate limit"
}

// AsFloodWait оборачивает ошибку gotd в доменную, если это FLOOD_WAIT.
func AsFloodWait(err error) (*FloodWait, bool) {
	var already *FloodWait
	if errors.As(err, &already) {
		return already, true
	}
	if d, ok := tgerr.AsFloodWait(err); ok {
		seconds := int(d / time.Second)
		if seconds < 1 {
			seconds = 1
		}
		return &FloodWait{RetryAfter: seconds}, true
	}
	return nil, false
}

// wrapTelegram приводит ошибку MTProto к доменной, сохраняя FLOOD_WAIT.
func wrapTelegram(err error, fallback error) error {
	if flood, ok := AsFloodWait(err); ok {
		return flood
	}
	return errors.Join(fallback, err)
}
