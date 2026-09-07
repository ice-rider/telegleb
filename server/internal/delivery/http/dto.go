package http

import (
	"strconv"
	"time"

	"telegleb/internal/core/domain"
)

// Все ключи — lowerCamelCase, на всех уровнях. Единая конвенция избавляет
// клиент от слоя переименования, который раньше расходился с сервером.

type meDTO struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
}

type entityDTO struct {
	Offset   int    `json:"offset"`
	Length   int    `json:"length"`
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	UserID   int64  `json:"userId,omitempty"`
	Language string `json:"language,omitempty"`
}

type mediaDTO struct {
	Ref      string `json:"ref"`
	Kind     string `json:"kind"`
	MimeType string `json:"mimeType,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	FileName string `json:"fileName,omitempty"`
}

type replyDTO struct {
	ID         int64  `json:"id"`
	SenderName string `json:"senderName,omitempty"`
	Text       string `json:"text,omitempty"`
}

type messageDTO struct {
	ID            int64       `json:"id"`
	ChatRef       string      `json:"chatRef"`
	SenderID      int64       `json:"senderId"`
	SenderName    string      `json:"senderName,omitempty"`
	Out           bool        `json:"out"`
	Text          string      `json:"text"`
	CreatedAt     string      `json:"createdAt"`
	EditedAt      string      `json:"editedAt,omitempty"`
	Entities      []entityDTO `json:"entities"`
	Media         *mediaDTO   `json:"media,omitempty"`
	ReplyTo       *replyDTO   `json:"replyTo,omitempty"`
	ForwardedFrom string      `json:"forwardedFrom,omitempty"`
	Views         int         `json:"views,omitempty"`
	Pinned        bool        `json:"pinned,omitempty"`
}

type chatDTO struct {
	Ref                 string      `json:"ref"`
	ID                  int64       `json:"id"`
	Type                string      `json:"type"`
	Title               string      `json:"title"`
	UnreadCount         int         `json:"unreadCount"`
	UnreadMentionsCount int         `json:"unreadMentionsCount"`
	MarkedUnread        bool        `json:"markedUnread"`
	Pinned              bool        `json:"pinned"`
	Muted               bool        `json:"muted"`
	Archived            bool        `json:"archived"`
	FolderIDs           []int       `json:"folderIds"`
	Order               int         `json:"order"`
	IsForum             bool        `json:"isForum"`
	LastMessage         *messageDTO `json:"lastMessage,omitempty"`
}

type topicDTO struct {
	ID                  int         `json:"id"`
	Title               string      `json:"title"`
	IconColor           int         `json:"iconColor,omitempty"`
	IconEmojiID         string      `json:"iconEmojiId,omitempty"`
	UnreadCount         int         `json:"unreadCount"`
	UnreadMentionsCount int         `json:"unreadMentionsCount"`
	Pinned              bool        `json:"pinned"`
	Closed              bool        `json:"closed"`
	Hidden              bool        `json:"hidden"`
	Order               int         `json:"order"`
	LastMessage         *messageDTO `json:"lastMessage,omitempty"`
}

type folderDTO struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Emoticon string `json:"emoticon,omitempty"`
	Order    int    `json:"order"`
}

// formatTime — единственный способ сериализовать время в проекте. Раскладка
// "2006-01-02T15:04:05Z" запрещена: буква Z в ней литерал, значение остаётся
// локальным, и клиент получает сдвиг на таймзону сервера.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func mapMe(m domain.Me) meDTO {
	return meDTO{
		ID:        m.ID,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Username:  m.Username,
		Phone:     m.Phone,
	}
}

func mapMessage(m domain.Message) messageDTO {
	entities := make([]entityDTO, 0, len(m.Entities))
	for _, e := range m.Entities {
		entities = append(entities, entityDTO{
			Offset:   e.Offset,
			Length:   e.Length,
			Type:     e.Type,
			URL:      e.URL,
			UserID:   e.UserID,
			Language: e.Language,
		})
	}

	dto := messageDTO{
		ID:            m.ID,
		ChatRef:       m.ChatRef,
		SenderID:      m.SenderID,
		SenderName:    m.SenderName,
		Out:           m.Out,
		Text:          m.Text,
		CreatedAt:     formatTime(m.CreatedAt),
		Entities:      entities,
		ForwardedFrom: m.ForwardedFrom,
		Views:         m.Views,
		Pinned:        m.Pinned,
	}

	// Отсутствующего времени не существует: поле просто опускается, а не
	// приезжает нулевой датой 0001-01-01.
	if m.EditedAt != nil {
		dto.EditedAt = formatTime(*m.EditedAt)
	}
	if m.Media != nil {
		dto.Media = &mediaDTO{
			Ref:      m.Media.Ref.String(),
			Kind:     string(m.Media.Kind),
			MimeType: m.Media.MimeType,
			Size:     m.Media.Size,
			Width:    m.Media.Width,
			Height:   m.Media.Height,
			FileName: m.Media.FileName,
		}
	}
	if m.ReplyTo != nil {
		dto.ReplyTo = &replyDTO{
			ID:         m.ReplyTo.ID,
			SenderName: m.ReplyTo.SenderName,
			Text:       m.ReplyTo.Text,
		}
	}
	return dto
}

func mapChat(c domain.Chat) chatDTO {
	folderIDs := c.FolderIDs
	if folderIDs == nil {
		folderIDs = []int{}
	}

	dto := chatDTO{
		Ref:                 c.Peer.Ref(),
		ID:                  c.Peer.ID,
		Type:                string(c.Type),
		Title:               c.Title,
		UnreadCount:         c.UnreadCount,
		UnreadMentionsCount: c.UnreadMentionsCount,
		MarkedUnread:        c.MarkedUnread,
		Pinned:              c.Pinned,
		Muted:               c.Muted,
		Archived:            c.Archived,
		FolderIDs:           folderIDs,
		Order:               c.Order,
		IsForum:             c.IsForum,
	}
	if c.LastMessage != nil {
		msg := mapMessage(*c.LastMessage)
		dto.LastMessage = &msg
	}
	return dto
}

func mapTopic(t domain.Topic) topicDTO {
	dto := topicDTO{
		ID:                  t.ID,
		Title:               t.Title,
		IconColor:           t.IconColor,
		UnreadCount:         t.UnreadCount,
		UnreadMentionsCount: t.UnreadMentionsCount,
		Pinned:              t.Pinned,
		Closed:              t.Closed,
		Hidden:              t.Hidden,
		Order:               t.Order,
	}
	// id кастомного эмодзи — int64 полного диапазона, в double он не влезает.
	if t.IconEmojiID != 0 {
		dto.IconEmojiID = strconv.FormatInt(t.IconEmojiID, 10)
	}
	if t.LastMessage != nil {
		msg := mapMessage(*t.LastMessage)
		dto.LastMessage = &msg
	}
	return dto
}

func mapFolder(f domain.Folder) folderDTO {
	return folderDTO{ID: f.ID, Title: f.Title, Emoticon: f.Emoticon, Order: f.Order}
}
