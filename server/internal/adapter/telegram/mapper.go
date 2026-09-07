package telegram

import (
	"time"

	"telegleb/internal/core/domain"

	"github.com/gotd/td/tg"
)

// msgKey адресует сообщение парой (чат, id). Ключа из одного msg.ID
// недостаточно: нумерация уникальна лишь внутри чата, у каналов она своя и
// начинается с единицы, поэтому плоская карта по id склеивает сообщения
// разных диалогов.
type msgKey struct {
	peer  string
	msgID int
}

// dicts — словари сущностей из ответа MTProto. Telegram отдаёт чаты, каналы и
// пользователей отдельными списками, а в диалогах и сообщениях ссылается на
// них по id.
type dicts struct {
	users map[int64]*tg.User
	chats map[int64]*tg.Chat
	chans map[int64]*tg.Channel
	ownID int64
	msgs  map[msgKey]*tg.Message
}

func newDicts(chats []tg.ChatClass, users []tg.UserClass, ownID int64) *dicts {
	d := &dicts{
		users: make(map[int64]*tg.User, len(users)),
		chats: make(map[int64]*tg.Chat),
		chans: make(map[int64]*tg.Channel),
		ownID: ownID,
		msgs:  make(map[msgKey]*tg.Message),
	}
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			d.users[user.ID] = user
		}
	}
	for _, c := range chats {
		switch v := c.(type) {
		case *tg.Chat:
			d.chats[v.ID] = v
		case *tg.Channel:
			d.chans[v.ID] = v
		}
	}
	return d
}

func (d *dicts) addMessages(messages []tg.MessageClass) {
	for _, m := range messages {
		msg, ok := m.(*tg.Message)
		if !ok {
			continue
		}
		if key, ok := peerClassKey(msg.PeerID); ok {
			d.msgs[msgKey{peer: key, msgID: msg.ID}] = msg
		}
	}
}

// peerOf восстанавливает полный адрес чата, доставая accessHash из словарей.
func (d *dicts) peerOf(p tg.PeerClass) (domain.Peer, domain.ChatType, string, bool) {
	switch v := p.(type) {
	case *tg.PeerUser:
		user, found := d.users[v.UserID]
		if !found {
			return domain.Peer{}, "", "", false
		}
		return domain.Peer{Kind: domain.PeerUser, ID: v.UserID, AccessHash: user.AccessHash},
			domain.ChatTypeDirect, userTitle(user, v.UserID == d.ownID), true

	case *tg.PeerChat:
		chat, found := d.chats[v.ChatID]
		if !found {
			return domain.Peer{}, "", "", false
		}
		return domain.Peer{Kind: domain.PeerChat, ID: v.ChatID},
			domain.ChatTypeGroup, chat.Title, true

	case *tg.PeerChannel:
		ch, found := d.chans[v.ChannelID]
		if !found {
			return domain.Peer{}, "", "", false
		}
		// Супергруппа с точки зрения пользователя — группа, а не канал.
		chatType := domain.ChatTypeChannel
		if ch.Megagroup {
			chatType = domain.ChatTypeGroup
		}
		return domain.Peer{Kind: domain.PeerChannel, ID: v.ChannelID, AccessHash: ch.AccessHash},
			chatType, ch.Title, true

	default:
		return domain.Peer{}, "", "", false
	}
}

func userTitle(u *tg.User, isSelf bool) string {
	if isSelf {
		return "Избранное"
	}
	name := u.FirstName
	if u.LastName != "" {
		if name != "" {
			name += " "
		}
		name += u.LastName
	}
	if name == "" {
		name = u.Username
	}
	if name == "" {
		name = "Удалённый аккаунт"
	}
	return name
}

func (d *dicts) senderName(p tg.PeerClass) string {
	switch v := p.(type) {
	case *tg.PeerUser:
		if u, ok := d.users[v.UserID]; ok {
			return userTitle(u, v.UserID == d.ownID)
		}
	case *tg.PeerChat:
		if c, ok := d.chats[v.ChatID]; ok {
			return c.Title
		}
	case *tg.PeerChannel:
		if c, ok := d.chans[v.ChannelID]; ok {
			return c.Title
		}
	}
	return ""
}

func (d *dicts) mapMessage(msg *tg.Message, peer domain.Peer) domain.Message {
	out := domain.Message{
		ID:        int64(msg.ID),
		ChatRef:   peer.Ref(),
		Out:       msg.Out,
		Text:      msg.Message,
		CreatedAt: time.Unix(int64(msg.Date), 0).UTC(),
		Views:     msg.Views,
		Pinned:    msg.Pinned,
		Entities:  mapEntities(msg.Entities),
	}

	if msg.FromID != nil {
		if u, ok := msg.FromID.(*tg.PeerUser); ok {
			out.SenderID = u.UserID
		}
		out.SenderName = d.senderName(msg.FromID)
	} else {
		// У каналов и в личных чатах отправитель не указан явно: это сам пир.
		out.SenderID = peer.ID
		out.SenderName = d.senderName(msg.PeerID)
	}

	if msg.EditDate != 0 {
		edited := time.Unix(int64(msg.EditDate), 0).UTC()
		out.EditedAt = &edited
	}

	if fwd, ok := msg.GetFwdFrom(); ok {
		out.ForwardedFrom = fwd.FromName
		if out.ForwardedFrom == "" && fwd.FromID != nil {
			out.ForwardedFrom = d.senderName(fwd.FromID)
		}
	}

	if msg.ReplyTo != nil {
		if h, ok := msg.ReplyTo.(*tg.MessageReplyHeader); ok && h.ReplyToMsgID != 0 {
			preview := domain.ReplyPreview{ID: int64(h.ReplyToMsgID)}
			if orig, ok := d.msgs[msgKey{peer: peerKey(peer.Kind, peer.ID), msgID: h.ReplyToMsgID}]; ok {
				preview.Text = orig.Message
				if orig.FromID != nil {
					preview.SenderName = d.senderName(orig.FromID)
				}
			}
			out.ReplyTo = &preview
		}
	}

	out.Media = mapMedia(msg, peer)
	return out
}

func mapMedia(msg *tg.Message, peer domain.Peer) *domain.Media {
	if msg.Media == nil {
		return nil
	}

	switch m := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := m.Photo.(*tg.Photo)
		if !ok {
			return nil
		}
		media := &domain.Media{Kind: domain.MediaPhoto, MimeType: "image/jpeg"}
		for _, size := range photo.Sizes {
			if s, ok := size.(*tg.PhotoSize); ok && s.W > media.Width {
				media.Width, media.Height, media.Size = s.W, s.H, int64(s.Size)
			}
		}
		media.Ref = domain.MediaRef{Peer: peer, MessageID: msg.ID, Kind: domain.MediaPhoto}
		return media

	case *tg.MessageMediaDocument:
		doc, ok := m.Document.(*tg.Document)
		if !ok {
			return nil
		}
		media := &domain.Media{
			Kind:     domain.MediaDocument,
			MimeType: doc.MimeType,
			Size:     doc.Size,
		}
		for _, attr := range doc.Attributes {
			switch a := attr.(type) {
			case *tg.DocumentAttributeFilename:
				media.FileName = a.FileName
			case *tg.DocumentAttributeSticker:
				media.Kind = domain.MediaSticker
			case *tg.DocumentAttributeAnimated:
				media.Kind = domain.MediaGIF
			case *tg.DocumentAttributeVideo:
				if media.Kind == domain.MediaDocument {
					media.Kind = domain.MediaVideo
				}
				media.Width, media.Height = a.W, a.H
			case *tg.DocumentAttributeAudio:
				if media.Kind == domain.MediaDocument {
					if a.Voice {
						media.Kind = domain.MediaVoice
					} else {
						media.Kind = domain.MediaAudio
					}
				}
			}
		}
		media.Ref = domain.MediaRef{Peer: peer, MessageID: msg.ID, Kind: media.Kind}
		return media

	default:
		return nil
	}
}

func mapEntities(entities []tg.MessageEntityClass) []domain.MessageEntity {
	result := make([]domain.MessageEntity, 0, len(entities))
	for _, e := range entities {
		var mapped domain.MessageEntity
		switch ent := e.(type) {
		case *tg.MessageEntityURL:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "url"}
		case *tg.MessageEntityTextURL:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "text_url", URL: ent.URL}
		case *tg.MessageEntityMentionName:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "mention_name", UserID: ent.UserID}
		case *tg.MessageEntityPre:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "pre", Language: ent.Language}
		case *tg.MessageEntityBold:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "bold"}
		case *tg.MessageEntityItalic:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "italic"}
		case *tg.MessageEntityUnderline:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "underline"}
		case *tg.MessageEntityStrike:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "strike"}
		case *tg.MessageEntitySpoiler:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "spoiler"}
		case *tg.MessageEntityCode:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "code"}
		case *tg.MessageEntityBlockquote:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "blockquote"}
		case *tg.MessageEntityMention:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "mention"}
		case *tg.MessageEntityHashtag:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "hashtag"}
		case *tg.MessageEntityBotCommand:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "bot_command"}
		case *tg.MessageEntityEmail:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "email"}
		case *tg.MessageEntityPhone:
			mapped = domain.MessageEntity{Offset: ent.Offset, Length: ent.Length, Type: "phone"}
		default:
			// Неизвестные серверу типы в контракт не попадают.
			continue
		}
		result = append(result, mapped)
	}
	return result
}
