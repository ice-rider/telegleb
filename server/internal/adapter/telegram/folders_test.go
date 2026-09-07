package telegram

import (
	"testing"

	"telegleb/internal/core/domain"

	"github.com/gotd/td/tg"
)

func dialog(kind domain.PeerKind, id int64, chatType domain.ChatType, opts ...func(*dialogInfo)) dialogInfo {
	info := dialogInfo{
		chat: domain.Chat{
			Peer:      domain.Peer{Kind: kind, ID: id},
			Type:      chatType,
			FolderIDs: []int{},
		},
		key: peerKey(kind, id),
	}
	for _, opt := range opts {
		opt(&info)
	}
	return info
}

func folderIDs(d dialogInfo) []int { return d.chat.FolderIDs }

func hasFolder(d dialogInfo, id int) bool {
	for _, f := range folderIDs(d) {
		if f == id {
			return true
		}
	}
	return false
}

// Папки вида «Личные» или «Каналы» держат include_peers пустым и наполняются
// исключительно предикатами. Учитывать только include_peers — значит всегда
// показывать такие папки пустыми.
func TestPredicatesFillFolderWithoutExplicitPeers(t *testing.T) {
	dialogs := []dialogInfo{
		dialog(domain.PeerUser, 1, domain.ChatTypeDirect, func(d *dialogInfo) { d.isContact = true }),
		dialog(domain.PeerUser, 2, domain.ChatTypeDirect),
		dialog(domain.PeerUser, 3, domain.ChatTypeDirect, func(d *dialogInfo) { d.isBot = true }),
		dialog(domain.PeerChannel, 4, domain.ChatTypeChannel),
		dialog(domain.PeerChat, 5, domain.ChatTypeGroup),
	}

	filters := []tg.DialogFilterClass{
		&tg.DialogFilter{ID: 10, Title: tg.TextWithEntities{Text: "Личные"}, Contacts: true},
		&tg.DialogFilter{ID: 11, Title: tg.TextWithEntities{Text: "Каналы"}, Broadcasts: true},
		&tg.DialogFilter{ID: 12, Title: tg.TextWithEntities{Text: "Группы"}, Groups: true},
		&tg.DialogFilter{ID: 13, Title: tg.TextWithEntities{Text: "Боты"}, Bots: true},
	}

	folders := assignFolders(dialogs, filters, 0)
	if len(folders) != 4 {
		t.Fatalf("ожидалось 4 папки, получено %d", len(folders))
	}

	if !hasFolder(dialogs[0], 10) {
		t.Error("контакт должен попасть в папку контактов")
	}
	if hasFolder(dialogs[1], 10) {
		t.Error("не-контакт не должен попасть в папку контактов")
	}
	if !hasFolder(dialogs[3], 11) {
		t.Error("канал должен попасть в папку каналов")
	}
	if !hasFolder(dialogs[4], 12) {
		t.Error("группа должна попасть в папку групп")
	}
	if !hasFolder(dialogs[2], 13) {
		t.Error("бот должен попасть в папку ботов")
	}
	if hasFolder(dialogs[2], 10) {
		t.Error("бот не должен считаться контактом")
	}
}

func TestPinnedPeersCountAsMembers(t *testing.T) {
	dialogs := []dialogInfo{dialog(domain.PeerChannel, 42, domain.ChatTypeChannel)}
	filters := []tg.DialogFilterClass{
		&tg.DialogFilter{
			ID:          20,
			Title:       tg.TextWithEntities{Text: "Работа"},
			PinnedPeers: []tg.InputPeerClass{&tg.InputPeerChannel{ChannelID: 42, AccessHash: 999}},
		},
	}

	assignFolders(dialogs, filters, 0)
	if !hasFolder(dialogs[0], 20) {
		t.Error("закреплённый в папке чат должен считаться её участником")
	}
}

// AccessHash в описании папки может отличаться от актуального, поэтому пиры
// сравниваются только по виду и id.
func TestPeersMatchIgnoringAccessHash(t *testing.T) {
	dialogs := []dialogInfo{dialog(domain.PeerUser, 7, domain.ChatTypeDirect)}
	dialogs[0].chat.Peer.AccessHash = 111

	filters := []tg.DialogFilterClass{
		&tg.DialogFilter{
			ID:           21,
			Title:        tg.TextWithEntities{Text: "Список"},
			IncludePeers: []tg.InputPeerClass{&tg.InputPeerUser{UserID: 7, AccessHash: 222}},
		},
	}

	assignFolders(dialogs, filters, 0)
	if !hasFolder(dialogs[0], 21) {
		t.Error("пиры должны сопоставляться без учёта accessHash")
	}
}

func TestExclusionsAreApplied(t *testing.T) {
	dialogs := []dialogInfo{
		dialog(domain.PeerChannel, 1, domain.ChatTypeChannel, func(d *dialogInfo) { d.chat.Muted = true }),
		dialog(domain.PeerChannel, 2, domain.ChatTypeChannel, func(d *dialogInfo) { d.chat.Archived = true }),
		dialog(domain.PeerChannel, 3, domain.ChatTypeChannel),
		dialog(domain.PeerChannel, 4, domain.ChatTypeChannel),
	}

	filters := []tg.DialogFilterClass{
		&tg.DialogFilter{
			ID:              30,
			Title:           tg.TextWithEntities{Text: "Каналы"},
			Broadcasts:      true,
			ExcludeMuted:    true,
			ExcludeArchived: true,
			ExcludePeers:    []tg.InputPeerClass{&tg.InputPeerChannel{ChannelID: 4}},
		},
	}

	assignFolders(dialogs, filters, 0)

	if hasFolder(dialogs[0], 30) {
		t.Error("exclude_muted должен исключать заглушённые чаты")
	}
	if hasFolder(dialogs[1], 30) {
		t.Error("exclude_archived должен исключать архивные чаты")
	}
	if hasFolder(dialogs[3], 30) {
		t.Error("exclude_peers должен исключать перечисленные чаты")
	}
	if !hasFolder(dialogs[2], 30) {
		t.Error("обычный канал должен остаться в папке")
	}
}

// Папки-ссылки приходят отдельным типом; отбрасывать их — значит терять
// вкладку целиком.
func TestChatlistFoldersAreSupported(t *testing.T) {
	dialogs := []dialogInfo{dialog(domain.PeerChannel, 5, domain.ChatTypeChannel)}
	filters := []tg.DialogFilterClass{
		&tg.DialogFilterChatlist{
			ID:           40,
			Title:        tg.TextWithEntities{Text: "Подборка"},
			IncludePeers: []tg.InputPeerClass{&tg.InputPeerChannel{ChannelID: 5}},
		},
	}

	folders := assignFolders(dialogs, filters, 0)
	if len(folders) != 1 || folders[0].Title != "Подборка" {
		t.Fatalf("папка-ссылка должна попасть в список: %+v", folders)
	}
	if !hasFolder(dialogs[0], 40) {
		t.Error("чат из папки-ссылки должен получить её id")
	}
}

// «Все чаты» приходит как dialogFilterDefault и отдельной вкладкой быть не должна.
func TestDefaultFilterIsNotAFolder(t *testing.T) {
	folders := assignFolders(nil, []tg.DialogFilterClass{&tg.DialogFilterDefault{}}, 0)
	if len(folders) != 0 {
		t.Errorf("dialogFilterDefault не должен превращаться в папку, получено %+v", folders)
	}
}

// «Избранное» приходит в папках как InputPeerSelf и должно резолвиться в
// собственный id пользователя.
func TestSelfPeerResolvesToOwnID(t *testing.T) {
	const ownID int64 = 777
	dialogs := []dialogInfo{dialog(domain.PeerUser, ownID, domain.ChatTypeDirect)}
	filters := []tg.DialogFilterClass{
		&tg.DialogFilter{
			ID:           50,
			Title:        tg.TextWithEntities{Text: "Избранное"},
			IncludePeers: []tg.InputPeerClass{&tg.InputPeerSelf{}},
		},
	}

	assignFolders(dialogs, filters, ownID)
	if !hasFolder(dialogs[0], 50) {
		t.Error("InputPeerSelf должен сопоставляться с собственным чатом")
	}
}
