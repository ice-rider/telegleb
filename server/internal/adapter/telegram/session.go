package telegram

import (
	"context"
	"sync"

	"github.com/gotd/td/session"
)

type SessionBridge struct {
	mu          sync.Mutex
	sessionData *[]byte
	onUpdate    func(ctx context.Context, data []byte) error
}

func NewSessionBridge(sessionData *[]byte, onUpdate func(ctx context.Context, data []byte) error) *SessionBridge {
	return &SessionBridge{
		sessionData: sessionData,
		onUpdate:    onUpdate,
	}
}

func (s *SessionBridge) LoadSession(ctx context.Context) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sessionData == nil || len(*s.sessionData) == 0 {
		return nil, session.ErrNotFound
	}
	return *s.sessionData, nil
}

func (s *SessionBridge) StoreSession(ctx context.Context, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.onUpdate != nil {
		return s.onUpdate(ctx, data)
	}
	return nil
}