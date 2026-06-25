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

func compileMessageMap(messages []tg.MessageClass) map[int]*tg.Message {
	msgMap := make(map[int]*tg.Message, len(messages))
	for _, m := range messages {
		if msg, ok := m.(*tg.Message); ok {
			msgMap[msg.ID] = msg
		}
	}
	return msgMap
}

func mapTelegramDialogs(dialogs []tg.DialogClass, chatMap map[int64]tg.ChatClass, userMap map[int64]tg.UserClass, msgMap map[int]*tg.Message) []domain.Chat {
	result := make([]domain.Chat, 0, len(dialogs))

	for _, dClass := range dialogs {
		d, ok := dClass.(*tg.Dialog)
		if !ok {
			continue
		}

		if d.FolderID == 1 {
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

		if topMsg, found := msgMap[d.TopMessage]; found {
			chat.LastMessage = domain.Message{
				ID:        int64(topMsg.ID),
				ChatID:    chat.ID,
				Text:      topMsg.Message,
				CreatedAt: time.Unix(int64(topMsg.Date), 0),
			}
		}

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

		editDate := time.Time{}
		if msg.EditDate != 0 {
			editDate = time.Unix(int64(msg.EditDate), 0)
		}

		var replyToMsgID, replyToPeer int64
		if msg.ReplyTo != nil {
			if replyHeader, ok := msg.ReplyTo.(*tg.MessageReplyHeader); ok {
				replyToMsgID = int64(replyHeader.ReplyToMsgID)
				if replyHeader.ReplyToPeerID != nil {
					switch p := replyHeader.ReplyToPeerID.(type) {
					case *tg.PeerUser:
						replyToPeer = p.UserID
					case *tg.PeerChat:
						replyToPeer = p.ChatID
					case *tg.PeerChannel:
						replyToPeer = p.ChannelID
					}
				}
			}
		}

		var fwdFromName string
		var fwdFromDate time.Time
		var fwdFromChannelID, fwdFromUserID int64
		if fwdFrom, ok := msg.GetFwdFrom(); ok {
			fwdFromName = fwdFrom.FromName
			if fwdFrom.Date != 0 {
				fwdFromDate = time.Unix(int64(fwdFrom.Date), 0)
			}
			if fwdFrom.FromID != nil {
				switch p := fwdFrom.FromID.(type) {
				case *tg.PeerUser:
					fwdFromUserID = p.UserID
				case *tg.PeerChannel:
					fwdFromChannelID = p.ChannelID
				}
			}
		}

		var repliesCount int
		var repliesMaxID int64
		if replies, ok := msg.GetReplies(); ok {
			repliesCount = replies.Replies
			repliesMaxID = int64(replies.MaxID)
		}

		entities := make([]domain.MessageEntity, 0, len(msg.Entities))
		for _, e := range msg.Entities {
			if e == nil {
				continue
			}
			switch ent := e.(type) {
			case *tg.MessageEntityURL:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "url",
				})
			case *tg.MessageEntityTextURL:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "text_url", URL: ent.URL,
				})
			case *tg.MessageEntityMentionName:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "mention_name", UserID: ent.UserID,
				})
			case *tg.MessageEntityBold:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "bold",
				})
			case *tg.MessageEntityItalic:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "italic",
				})
			case *tg.MessageEntityCode:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "code",
				})
			case *tg.MessageEntityPre:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "pre", URL: ent.Language,
				})
			case *tg.MessageEntityUnderline:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "underline",
				})
			case *tg.MessageEntityStrike:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "strike",
				})
			case *tg.MessageEntitySpoiler:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "spoiler",
				})
			case *tg.MessageEntityBlockquote:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "blockquote",
				})
			case *tg.MessageEntityMention:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "mention",
				})
			case *tg.MessageEntityHashtag:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "hashtag",
				})
			case *tg.MessageEntityBotCommand:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "bot_command",
				})
			case *tg.MessageEntityEmail:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "email",
				})
			case *tg.MessageEntityPhone:
				entities = append(entities, domain.MessageEntity{
					Offset: ent.Offset, Length: ent.Length, Type: "phone",
				})
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

			Out:        msg.Out,
			Mentioned:  msg.Mentioned,
			Silent:     msg.Silent,
			Post:       msg.Post,
			Pinned:     msg.Pinned,
			Noforwards: msg.Noforwards,
			EditDate:   editDate,
			Views:      msg.Views,
			Forwards:   msg.Forwards,
			GroupedID:  msg.GroupedID,
			ViaBotID:   msg.ViaBotID,
			PostAuthor: msg.PostAuthor,
			TTLPeriod:  msg.TTLPeriod,

			ReplyToMsgID: replyToMsgID,
			ReplyToPeer:  replyToPeer,

			FwdFromName:      fwdFromName,
			FwdFromDate:      fwdFromDate,
			FwdFromChannelID: fwdFromChannelID,
			FwdFromUserID:    fwdFromUserID,

			RepliesCount: repliesCount,
			RepliesMaxID: repliesMaxID,

			Entities: entities,
		})
	}

	return result
}
