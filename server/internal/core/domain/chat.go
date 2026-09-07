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
	// IsForum — супергруппа с темами. У такого чата нет плоской истории:
	// сообщения живут внутри тем, поэтому клиент сначала показывает список тем.
	IsForum bool
}

// Topic — тема внутри форума-супергруппы. По сути это отдельный чат, у
// которого есть свои непрочитанные и своя история, но живёт он внутри
// родительского чата и адресуется его ID.
type Topic struct {
	ID                  int
	Title               string
	IconColor           int
	IconEmojiID         int64
	UnreadCount         int
	UnreadMentionsCount int
	Pinned              bool
	Closed              bool
	Hidden              bool
	Order               int
	LastMessage         *Message
}

// Dashboard — атомарный снимок: чаты, папки и профиль приезжают одним ответом.
type Dashboard struct {
	Me      Me
	Chats   []Chat
	Folders []Folder
	// Truncated — список диалогов упёрся в потолок. Раскладка по папкам в
	// этом случае неполна, и клиенту стоит об этом сказать.
	Truncated bool
	Stats     DashboardStats
}

// DashboardStats — диагностика загрузки. Нужна, чтобы «в папке 2 чата вместо
// 13» можно было объяснить числами, а не догадками.
type DashboardStats struct {
	Pages              int
	RawDialogs         int
	SkippedUnknownPeer int
	SkippedNotDialog   int
}
