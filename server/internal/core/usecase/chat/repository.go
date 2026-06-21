package chat

import (
	"context"

	"telegleb/internal/core/domain"
)

type ChatRepository interface {
	GetChats(ctx context.Context, sessionToken string, limit int, offset int) ([]domain.Chat, error)
	GetFolders(ctx context.Context, sessionToken string) ([]domain.Folder, error)
}