package domain

type ChatType string

const (
	ChatTypeDirect  ChatType = "direct"
	ChatTypeGroup   ChatType = "group"
	ChatTypeChannel ChatType = "channel"
)

// Chat — диалог в том виде, в котором его рисует клиент.
//
// Archived и FolderIDs резолвит сервер, и они лежат на самом чате, а не в
// отдельном справочнике папок. Это сделано намеренно: клиент не может
// отрисовать чат, не зная его размещения, потому что размещение приезжает тем
// же объектом. Схема «список папок отдельно, чаты отдельно» допускает кадр, в
// котором чаты уже есть, а раскладка по папкам ещё нет — именно так архивные
// чаты мелькают в общем списке.
type Chat struct {
	Peer                Peer
	Type                ChatType
	Title               string
	UnreadCount         int
	UnreadMentionsCount int
	MarkedUnread        bool
	Pinned              bool
	Muted               bool
	Archived            bool
	FolderIDs           []int
	Order               int
	LastMessage         *Message
}

// Dashboard — атомарный снимок: чаты, папки и профиль приезжают одним ответом.
type Dashboard struct {
	Me         Me
	Chats      []Chat
	Folders    []Folder
	NextCursor string
}
