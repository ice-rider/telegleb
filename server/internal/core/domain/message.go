package domain

import "time"

type MessageEntity struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
	URL    string `json:"url,omitempty"`
	UserID int64  `json:"userId,omitempty"`
}

type Message struct {
	ID        int64
	ChatID    int64
	SenderID  int64
	Text      string
	CreatedAt time.Time
	HasMedia  bool
	MediaId   string

	Out        bool
	Mentioned  bool
	Silent     bool
	Post       bool
	Pinned     bool
	Noforwards bool
	EditDate   time.Time
	Views      int
	Forwards   int
	GroupedID  int64
	ViaBotID   int64
	PostAuthor string
	TTLPeriod  int

	ReplyToMsgID int64
	ReplyToPeer  int64

	FwdFromName      string
	FwdFromDate      time.Time
	FwdFromChannelID int64
	FwdFromUserID    int64

	RepliesCount   int
	RepliesMaxID   int64

	Entities []MessageEntity
}
