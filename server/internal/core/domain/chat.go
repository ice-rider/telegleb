package domain

type ChatType int

const (
	DIRECT ChatType = iota
	GROUP
	CHANNEL
)

type Chat struct {
	ID          int64
	Title       string
	Type        ChatType
	UnreadCount int
	LastMessage Message
}
