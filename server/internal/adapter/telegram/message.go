package telegram

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/gotd/td/tg"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/message"
)

type TelegramMessageRepository struct {
	adapter *TelegramAdapter
}

func NewTelegramMessageRepository(adapter *TelegramAdapter) message.MessageRepository {
	return &TelegramMessageRepository{adapter: adapter}
}

func (r *TelegramMessageRepository) GetChatHistory(ctx context.Context, sessionToken string, chatIDStr string, limit int, offsetID int) ([]domain.Message, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat id format: %w", err)
	}

	resp, err := client.API.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:     &tg.InputPeerUser{UserID: chatID},
		Limit:    limit,
		OffsetID: offsetID,
	})
	if err != nil {
		// Реактивный фолбэк для каналов/супергрупп
		resp, err = client.API.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     &tg.InputPeerChannel{ChannelID: chatID},
			Limit:    limit,
			OffsetID: offsetID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch chat history: %w", err)
		}
	}

	switch res := resp.(type) {
	case *tg.MessagesMessagesSlice:
		msgs := mapTelegramMessages(res.Messages)
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
		})
		return msgs, nil
	case *tg.MessagesMessages:
		msgs := mapTelegramMessages(res.Messages)
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
		})
		return msgs, nil
	case *tg.MessagesChannelMessages:
		msgs := mapTelegramMessages(res.Messages)
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
		})
		return msgs, nil
	default:
		return []domain.Message{}, nil
	}
}

func (r *TelegramMessageRepository) SendMessage(ctx context.Context, sessionToken string, chatIDStr string, text string) (domain.Message, error) {
	var emptyMsg domain.Message
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return emptyMsg, err
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return emptyMsg, fmt.Errorf("invalid chat id format: %w", err)
	}

	randomID := rand.Int63()
	req := &tg.MessagesSendMessageRequest{
		Peer:     &tg.InputPeerUser{UserID: chatID},
		Message:  text,
		RandomID: randomID,
	}

	// Выполняем запрос к Telegram API
	updates, err := client.API.MessagesSendMessage(ctx, req)
	if err != nil {
		// Исправлено: Если это группа/канал, подменяем получателя в структуре req и отправляем ее же
		req.Peer = &tg.InputPeerChannel{ChannelID: chatID}
		updates, err = client.API.MessagesSendMessage(ctx, req)
		if err != nil {
			return emptyMsg, fmt.Errorf("failed to send telegram message: %w", err)
		}
	}

	// Безопасный фолбэк для ID на случай, если структура ответов пуста
	generatedID := int64(randomID & 0x7FFFFFFF)

	if updates != nil {
		switch u := updates.(type) {
		case *tg.UpdateShortSentMessage: // Исправлено: Update вместо Updates (в единственном числе)
			generatedID = int64(u.ID)
		case *tg.Updates:
			for _, updateClass := range u.Updates {
				if m, ok := updateClass.(*tg.UpdateNewMessage); ok {
					if msgObj, ok := m.Message.(*tg.Message); ok {
						generatedID = int64(msgObj.ID)
						break
					}
				}
				if m, ok := updateClass.(*tg.UpdateNewChannelMessage); ok {
					if msgObj, ok := m.Message.(*tg.Message); ok {
						generatedID = int64(msgObj.ID)
						break
					}
				}
			}
		}
	}

	return domain.Message{
		ID:        generatedID,
		ChatID:    chatID,
		SenderID:  0, // Идентификатор "селф"
		Text:      text,
		CreatedAt: time.Now(),
		HasMedia:  false,
		MediaId:   "",
	}, nil
}
