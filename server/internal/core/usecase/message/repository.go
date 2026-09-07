package message

import (
	"context"

	"telegleb/internal/core/domain"
)

type MessageRepository interface {
	// GetHistory возвращает сообщения по возрастанию ID. beforeID == 0 — свежая выдача.
	GetHistory(ctx context.Context, sessionToken string, peer domain.Peer, limit int, beforeID int) ([]domain.Message, error)
	// Send отправляет текст. randomID — ключ идемпотентности от клиента.
	Send(ctx context.Context, sessionToken string, peer domain.Peer, text string, randomID int64) (domain.Message, error)
}
