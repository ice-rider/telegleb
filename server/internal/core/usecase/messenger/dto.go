package messenger

import (
	"hash/fnv"
	"strings"

	"telegleb/internal/core/domain"
)

const (
	defaultPageLimit = 30
	maxPageLimit     = 100
)

// clampLimit защищает и от нуля (Telegram вернул бы пусто), и от запроса
// заведомо большей страницы, чем отдаёт API.
func clampLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultPageLimit
	case limit > maxPageLimit:
		return maxPageLimit
	default:
		return limit
	}
}

type LoadDashboardInput struct {
	SessionToken string
	Limit        int
	Cursor       string
}

type LoadDashboardOutput struct {
	Dashboard domain.Dashboard
}

type OpenChatInput struct {
	SessionToken string
	ChatRef      string
	// TopicID > 0 — история внутри темы форума.
	TopicID  int
	Limit    int
	BeforeID int
}

func (i OpenChatInput) Peer() (domain.Peer, error) {
	peer, err := domain.ParsePeer(i.ChatRef)
	if err != nil {
		return domain.Peer{}, ErrInvalidPeer
	}
	return peer, nil
}

type OpenChatOutput struct {
	Messages     []domain.Message
	NextBeforeID int
}

type SendMessageInput struct {
	SessionToken string
	ChatRef      string
	TopicID      int
	Text         string
	RandomID     string
}

func (i SendMessageInput) Validate() error {
	if strings.TrimSpace(i.Text) == "" {
		return ErrEmptyMessage
	}
	return nil
}

func (i SendMessageInput) Peer() (domain.Peer, error) {
	peer, err := domain.ParsePeer(i.ChatRef)
	if err != nil {
		return domain.Peer{}, ErrInvalidPeer
	}
	return peer, nil
}

// randomID превращает клиентский ключ идемпотентности в int64, который ждёт
// MTProto. Одинаковый ключ даёт одинаковое число, поэтому повторная отправка
// не создаёт дубль.
func (i SendMessageInput) randomID() int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(i.RandomID))
	return int64(h.Sum64() >> 1)
}

type SendMessageOutput struct {
	Message domain.Message
}

type StreamMediaInput struct {
	SessionToken string
	MediaRef     string
	Offset       int64
	Limit        int
}

func (i StreamMediaInput) Validate() error {
	if i.Offset < 0 || i.Limit <= 0 {
		return ErrInvalidRange
	}
	return nil
}

func (i StreamMediaInput) Ref() (domain.MediaRef, error) {
	ref, err := domain.ParseMediaRef(i.MediaRef)
	if err != nil {
		return domain.MediaRef{}, ErrInvalidMediaRef
	}
	return ref, nil
}

type ListTopicsInput struct {
	SessionToken string
	ChatRef      string
}

func (i ListTopicsInput) Peer() (domain.Peer, error) {
	peer, err := domain.ParsePeer(i.ChatRef)
	if err != nil {
		return domain.Peer{}, ErrInvalidPeer
	}
	return peer, nil
}

type ListTopicsOutput struct {
	Topics []domain.Topic
}
