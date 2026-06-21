package message

import (
	"context"
	"telegleb/internal/core/domain"
)

type MessageRepository interface {
	GetChatHistory(ctx context.Context, sessionToken string, chatID string, limit int, offset int) ([]domain.Message, error)
	SendMessage(ctx context.Context, sessionToken string, chatID string, content string) (domain.Message, error)
}