package chat

import (
	"context"

	"telegleb/internal/core/domain"
)

type ChatRepository interface {
	// GetDashboard отдаёт чаты, папки и профиль одним снимком. Это один метод,
	// а не три, потому что принадлежность чата к папке вычисляется по составу
	// самого списка диалогов (предикаты contacts/groups/broadcasts и т.д.), и
	// разъезжаться эти данные не должны даже теоретически.
	GetDashboard(ctx context.Context, sessionToken string, limit int, cursor string) (domain.Dashboard, error)
}
