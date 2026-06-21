package telegram

import (
	"fmt"
	"time"

	"telegleb/internal/core/domain"

	"github.com/gotd/td/tg"
)

func compileDictionaries(chats []tg.ChatClass, users []tg.UserClass) (map[int64]tg.ChatClass, map[int64]tg.UserClass) {
	chatMap := make(map[int64]tg.ChatClass, len(chats))
	for _, c := range chats {
		if c != nil {
			chatMap[c.GetID()] = c
		}
	}
	userMap := make(map[int64]tg.UserClass, len(users))
	for _, u := range users {
		if u != nil {
			userMap[u.GetID()] = u
		}
	}
	return chatMap, userMap
}

func mapTelegramDialogs(dialogs []tg.DialogClass, chatMap map[int64]tg.ChatClass, userMap map[int64]tg.UserClass) []domain.Chat {
	result := make([]domain.Chat, 0, len(dialogs))

	for _, dClass := range dialogs {
		d, ok := dClass.(*tg.Dialog)
		if !ok {
			continue
		}

		var chat domain.Chat
		peer := d.Peer

		switch p := peer.(type) {
		case *tg.PeerUser:
			chat.ID = p.UserID
			chat.Type = domain.DIRECT
			if u, found := userMap[p.UserID]; found {
				if userObj, ok := u.(*tg.User); ok {
					chat.Title = userObj.FirstName
					if userObj.LastName != "" {
						chat.Title += " " + userObj.LastName
					}
				}
			}
		case *tg.PeerChat:
			chat.ID = p.ChatID
			chat.Type = domain.GROUP
			if c, found := chatMap[p.ChatID]; found {
				if chatObj, ok := c.(*tg.Chat); ok {
					chat.Title = chatObj.Title
				}
			}
		case *tg.PeerChannel:
			chat.ID = p.ChannelID
			chat.Type = domain.CHANNEL
			if c, found := chatMap[p.ChannelID]; found {
				if channelObj, ok := c.(*tg.Channel); ok {
					chat.Title = channelObj.Title
					if channelObj.Megagroup {
						chat.Type = domain.GROUP
					}
				}
			}
		default:
			continue
		}

		chat.UnreadCount = d.UnreadCount
		result = append(result, chat)
	}

	return result
}

func mapTelegramMessages(messages []tg.MessageClass) []domain.Message {
	result := make([]domain.Message, 0, len(messages))

	for _, mClass := range messages {
		msg, ok := mClass.(*tg.Message)
		if !ok {
			continue
		}

		var chatID int64
		switch p := msg.PeerID.(type) {
		case *tg.PeerUser:
			chatID = p.UserID
		case *tg.PeerChat:
			chatID = p.ChatID
		case *tg.PeerChannel:
			chatID = p.ChannelID
		}

		var senderID int64
		if msg.FromID != nil {
			switch s := msg.FromID.(type) {
			case *tg.PeerUser:
				senderID = s.UserID
			}
		} else if chatID != 0 {
			senderID = chatID
		}

		hasMedia := false
		mediaID := ""
		if msg.Media != nil {
			hasMedia = true
			if document, ok := msg.Media.(*tg.MessageMediaDocument); ok {
				if doc, ok := document.Document.(*tg.Document); ok {
					mediaID = fmt.Sprintf("doc_%d_%d", doc.ID, doc.AccessHash)
				}
			} else if photo, ok := msg.Media.(*tg.MessageMediaPhoto); ok {
				if p, ok := photo.Photo.(*tg.Photo); ok {
					mediaID = fmt.Sprintf("photo_%d_%d", p.ID, p.AccessHash)
				}
			}
		}

		result = append(result, domain.Message{
			ID:        int64(msg.ID),
			ChatID:    chatID,
			SenderID:  senderID,
			Text:      msg.Message,
			CreatedAt: time.Unix(int64(msg.Date), 0),
			HasMedia:  hasMedia,
			MediaId:   mediaID,
		})
	}

	return result
}
