package telegram

import (
	"context"
	"fmt"
	"sort"
	"time"

	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/message"

	"github.com/gotd/td/tg"
)

type TelegramMessageRepository struct {
	adapter *TelegramAdapter
}

func NewTelegramMessageRepository(adapter *TelegramAdapter) message.MessageRepository {
	return &TelegramMessageRepository{adapter: adapter}
}

func (r *TelegramMessageRepository) GetHistory(ctx context.Context, sessionToken string, peer domain.Peer, limit int, beforeID int) ([]domain.Message, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// Тип пира известен из ref — перебирать InputPeer вслепую не нужно.
	resp, err := client.API.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:     inputPeer(peer),
		Limit:    limit,
		OffsetID: beforeID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat history: %w", err)
	}

	msgs, chats, users := unpackMessages(resp)
	d := newDicts(chats, users, 0)
	d.addMessages(msgs)

	result := make([]domain.Message, 0, len(msgs))
	for _, mClass := range msgs {
		msg, ok := mClass.(*tg.Message)
		if !ok {
			continue
		}
		result = append(result, d.mapMessage(msg, peer))
	}

	// Клиент рисует историю от старых к новым, Telegram отдаёт наоборот.
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (r *TelegramMessageRepository) Send(ctx context.Context, sessionToken string, peer domain.Peer, text string, randomID int64) (domain.Message, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return domain.Message{}, err
	}

	updates, err := client.API.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     inputPeer(peer),
		Message:  text,
		RandomID: randomID,
	})
	if err != nil {
		return domain.Message{}, fmt.Errorf("failed to send telegram message: %w", err)
	}

	return messageFromUpdates(updates, peer, text), nil
}

// messageFromUpdates достаёт из ответа настоящее сообщение. Возвращать
// заглушку с нулевым отправителем нельзя: клиент выравнивает пузырь по Out, и
// собственное сообщение уехало бы влево.
func messageFromUpdates(updates tg.UpdatesClass, peer domain.Peer, text string) domain.Message {
	fallback := domain.Message{
		ChatRef:   peer.Ref(),
		Out:       true,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}

	switch u := updates.(type) {
	case *tg.UpdateShortSentMessage:
		fallback.ID = int64(u.ID)
		fallback.CreatedAt = time.Unix(int64(u.Date), 0).UTC()
		fallback.Out = true
		fallback.Entities = mapEntities(u.Entities)
		return fallback

	case *tg.Updates:
		d := newDicts(u.Chats, u.Users, 0)
		for _, updClass := range u.Updates {
			var msgClass tg.MessageClass
			switch upd := updClass.(type) {
			case *tg.UpdateNewMessage:
				msgClass = upd.Message
			case *tg.UpdateNewChannelMessage:
				msgClass = upd.Message
			default:
				continue
			}
			if msg, ok := msgClass.(*tg.Message); ok {
				mapped := d.mapMessage(msg, peer)
				mapped.Out = true
				return mapped
			}
		}
	}

	return fallback
}

func unpackMessages(resp tg.MessagesMessagesClass) ([]tg.MessageClass, []tg.ChatClass, []tg.UserClass) {
	switch res := resp.(type) {
	case *tg.MessagesMessages:
		return res.Messages, res.Chats, res.Users
	case *tg.MessagesMessagesSlice:
		return res.Messages, res.Chats, res.Users
	case *tg.MessagesChannelMessages:
		return res.Messages, res.Chats, res.Users
	default:
		return nil, nil, nil
	}
}
