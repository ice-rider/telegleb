package domain

// Folder описывает только саму вкладку. Состав папки живёт в Chat.FolderIDs —
// см. комментарий к Chat.
type Folder struct {
	ID       int
	Title    string
	Emoticon string
	Order    int
}
