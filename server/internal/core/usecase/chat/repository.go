package chat

import (
	"context"

	"telegleb/internal/core/domain"
)

type ChatRepository interface {
	// GetDashboard отдаёт чаты, папки и профиль одним снимком. Это один метод,
	// а не три, потому что принадлежность чата к папке вычисляется по составу
	// всего списка диалогов, и разъезжаться эти данные не должны.
	GetDashboard(ctx context.Context, sessionToken string, limit int, cursor string) (domain.Dashboard, error)
	// GetTopics возвращает темы форума-супергруппы.
	GetTopics(ctx context.Context, sessionToken string, peer domain.Peer) ([]domain.Topic, error)
}
