package message

import (
	"context"

	"telegleb/internal/core/domain"
)

type MessageRepository interface {
	// GetHistory возвращает сообщения по возрастанию ID.
	// topicID > 0 — история внутри темы форума, 0 — обычная история чата.
	// beforeID == 0 — свежая выдача.
	GetHistory(ctx context.Context, sessionToken string, peer domain.Peer, topicID, limit, beforeID int) ([]domain.Message, error)
	// Send отправляет текст. topicID > 0 адресует сообщение в тему форума.
	// randomID — ключ идемпотентности от клиента.
	Send(ctx context.Context, sessionToken string, peer domain.Peer, topicID int, text string, randomID int64) (domain.Message, error)
}
