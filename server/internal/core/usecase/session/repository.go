package session

import (
	"context"
	"telegleb/internal/core/domain"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, session *domain.AuthSession) error
	GetSessionByToken(ctx context.Context, token string) (*domain.AuthSession, error)
	UpdateSession(ctx context.Context, session *domain.AuthSession) error
	DeleteSession(ctx context.Context, token string) error
}