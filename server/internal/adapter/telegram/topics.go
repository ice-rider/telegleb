package telegram

import (
	"context"
	"fmt"

	"telegleb/internal/core/domain"

	"github.com/gotd/td/tg"
)

const (
	topicPageSize = 100
	maxTopics     = 500
)

// GetTopics забирает темы форума целиком, страницами. Так же, как со списком
// диалогов: частичная выборка означала бы, что часть тем просто исчезла из
// интерфейса без всякого объяснения.
func (r *TelegramChatRepository) GetTopics(ctx context.Context, sessionToken string, peer domain.Peer) ([]domain.Topic, error) {
	if peer.Kind != domain.PeerChannel {
		return nil, fmt.Errorf("%w: темы бывают только у супергрупп", domain.ErrInvalidPeerRef)
	}

	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	var (
		topics     []domain.Topic
		offsetDate int
		offsetID   int
		offsetTop  int
	)

	for len(topics) < maxTopics {
		res, err := client.API.MessagesGetForumTopics(ctx, &tg.MessagesGetForumTopicsRequest{
			Peer:        inputPeer(peer),
			Limit:       topicPageSize,
			OffsetDate:  offsetDate,
			OffsetID:    offsetID,
			OffsetTopic: offsetTop,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch forum topics: %w", err)
		}

		d := newDicts(res.Chats, res.Users, 0)
		d.addMessages(res.Messages)

		before := len(topics)
		var lastTopic *tg.ForumTopic

		for _, tClass := range res.Topics {
			// ForumTopicDeleted приходит в выдаче и означает лишь то, что
			// темы больше нет; показывать её нечего.
			t, ok := tClass.(*tg.ForumTopic)
			if !ok {
				continue
			}
			lastTopic = t

			topic := domain.Topic{
				ID:                  t.ID,
				Title:               t.Title,
				IconColor:           t.IconColor,
				IconEmojiID:         t.IconEmojiID,
				UnreadCount:         t.UnreadCount,
				UnreadMentionsCount: t.UnreadMentionsCount,
				Pinned:              t.Pinned,
				Closed:              t.Closed,
				Hidden:              t.Hidden,
				Order:               len(topics),
			}
			if msg, found := d.msgs[msgKey{peer: peerKey(peer.Kind, peer.ID), msgID: t.TopMessage}]; found {
				mapped := d.mapMessage(msg, peer)
				topic.LastMessage = &mapped
			}
			topics = append(topics, topic)
		}

		// Выдача исчерпана либо страница ничего не добавила — выходим, чтобы
		// не крутиться на месте.
		if len(topics) == before || len(res.Topics) < topicPageSize || lastTopic == nil {
			break
		}
		offsetDate, offsetID, offsetTop = lastTopic.Date, lastTopic.TopMessage, lastTopic.ID
	}

	return topics, nil
}
