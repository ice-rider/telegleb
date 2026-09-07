package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type MessageEntity struct {
	Offset   int
	Length   int
	Type     string
	URL      string
	UserID   int64
	Language string
}

type MediaKind string

const (
	MediaPhoto    MediaKind = "photo"
	MediaVideo    MediaKind = "video"
	MediaAudio    MediaKind = "audio"
	MediaVoice    MediaKind = "voice"
	MediaDocument MediaKind = "document"
	MediaSticker  MediaKind = "sticker"
	MediaGIF      MediaKind = "gif"
)

var ErrInvalidMediaRef = errors.New("invalid media ref")

// MediaRef адресует вложение не по (id, accessHash) файла, а по сообщению, в
// котором оно лежит. Telegram протухает fileReference, поэтому сохранять его в
// идентификаторе бессмысленно — сервер перезапрашивает сообщение и достаёт
// свежий reference в момент скачивания.
type MediaRef struct {
	Peer      Peer
	MessageID int
	Kind      MediaKind
}

// String собирает ссылку вида "channel:123:-456~78901~photo". Разделитель "~",
// а не "/", чтобы ref оставался одним сегментом URL.
func (m MediaRef) String() string {
	return m.Peer.Ref() + "~" + strconv.Itoa(m.MessageID) + "~" + string(m.Kind)
}

func ParseMediaRef(s string) (MediaRef, error) {
	parts := strings.Split(s, "~")
	if len(parts) != 3 {
		return MediaRef{}, ErrInvalidMediaRef
	}

	peer, err := ParsePeer(parts[0])
	if err != nil {
		return MediaRef{}, ErrInvalidMediaRef
	}

	msgID, err := strconv.Atoi(parts[1])
	if err != nil {
		return MediaRef{}, ErrInvalidMediaRef
	}

	return MediaRef{Peer: peer, MessageID: msgID, Kind: MediaKind(parts[2])}, nil
}

type Media struct {
	Ref      MediaRef
	Kind     MediaKind
	MimeType string
	Size     int64
	Width    int
	Height   int
	FileName string
}

type ReplyPreview struct {
	ID         int64
	SenderName string
	Text       string
}

// Message намеренно узкий: поле появляется здесь только тогда, когда его
// кто-то рисует. Раздутый DTO — это трафик на каждом из сотни чатов дашборда
// и лишняя поверхность для рассинхрона контракта.
type Message struct {
	ID            int64
	ChatRef       string
	SenderID      int64
	SenderName    string
	Out           bool
	Text          string
	CreatedAt     time.Time
	EditedAt      *time.Time
	Entities      []MessageEntity
	Media         *Media
	ReplyTo       *ReplyPreview
	ForwardedFrom string
	Views         int
	Pinned        bool
}
