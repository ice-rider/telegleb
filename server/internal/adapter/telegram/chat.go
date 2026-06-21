package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"telegleb/internal/core/usecase/chat"
	"telegleb/internal/core/domain"
)

type TelegramChatRepository struct {
	adapter *TelegramAdapter
}

func NewTelegramChatRepository(adapter *TelegramAdapter) chat.ChatRepository {
	return &TelegramChatRepository{adapter: adapter}
}

func (r *TelegramChatRepository) GetChats(ctx context.Context, sessionToken string, limit int, offset int) ([]domain.Chat, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	resp, err := client.API.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetID:   offset,
		Limit:      limit,
		OffsetPeer: &tg.InputPeerEmpty{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch telegram dialogs: %w", err)
	}

	switch res := resp.(type) {
	case *tg.MessagesDialogsSlice:
		chatMap, userMap := compileDictionaries(res.Chats, res.Users)
		return mapTelegramDialogs(res.Dialogs, chatMap, userMap), nil
	case *tg.MessagesDialogs:
		chatMap, userMap := compileDictionaries(res.Chats, res.Users)
		return mapTelegramDialogs(res.Dialogs, chatMap, userMap), nil
	default:
		return []domain.Chat{}, nil
	}
}

func (r *TelegramChatRepository) GetFolders(ctx context.Context, sessionToken string) ([]domain.Folder, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	dialogFilters, err := client.API.MessagesGetDialogFilters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch telegram folders: %w", err)
	}

	var folders []domain.Folder
	for _, fClass := range dialogFilters.Filters {
		f, ok := fClass.(*tg.DialogFilter)
		if !ok {
			continue
		}

		var chatIDs []int64
		for _, peerClass := range f.IncludePeers {
			// Исправлено: переключаемся по типам InputPeerClass
			switch p := peerClass.(type) {
			case *tg.InputPeerUser:
				chatIDs = append(chatIDs, p.UserID)
			case *tg.InputPeerChat:
				chatIDs = append(chatIDs, p.ChatID)
			case *tg.InputPeerChannel:
				chatIDs = append(chatIDs, p.ChannelID)
			}
		}

		folders = append(folders, domain.Folder{
			ID:      f.ID,
			Title:   f.Title.Text,
			ChatIDs: chatIDs,
		})
	}

	return folders, nil
}
