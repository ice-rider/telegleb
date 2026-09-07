package telegram

import (
	"telegleb/internal/core/domain"

	"github.com/gotd/td/tg"
)

// assignFolders проставляет каждому чату список папок, в которые он входит, и
// возвращает описания самих папок.
//
// Принадлежность считается здесь, а не на клиенте, и записывается в сам чат:
// клиент не должен соединять список чатов со списком папок, потому что во
// время такого соединения существует кадр, где чаты уже есть, а раскладка ещё
// нет — тогда архивные и чужие для вкладки чаты успевают отрисоваться.
//
// Учитывается всё, чем Telegram задаёт папку: явные включения (pinned_peers,
// include_peers), предикаты по типу чата и исключения. Одних include_peers
// недостаточно — папки вида «Личные» или «Каналы» держат этот список пустым и
// наполняются исключительно предикатами.
func assignFolders(dialogs []dialogInfo, filters []tg.DialogFilterClass, ownID int64) []domain.Folder {
	folders := make([]domain.Folder, 0, len(filters))

	for _, fClass := range filters {
		switch f := fClass.(type) {
		case *tg.DialogFilter:
			folders = append(folders, domain.Folder{
				ID:       f.ID,
				Title:    f.Title.Text,
				Emoticon: f.Emoticon,
				Order:    len(folders),
			})

			include := peerKeySet(ownID, f.PinnedPeers, f.IncludePeers)
			exclude := peerKeySet(ownID, f.ExcludePeers)

			for i := range dialogs {
				d := &dialogs[i]
				if _, excluded := exclude[d.key]; excluded {
					continue
				}
				if _, included := include[d.key]; included {
					d.chat.FolderIDs = append(d.chat.FolderIDs, f.ID)
					continue
				}
				if !matchesPredicates(f, d) {
					continue
				}
				if f.ExcludeMuted && d.chat.Muted {
					continue
				}
				if f.ExcludeRead && d.chat.UnreadCount == 0 && !d.chat.MarkedUnread {
					continue
				}
				if f.ExcludeArchived && d.chat.Archived {
					continue
				}
				d.chat.FolderIDs = append(d.chat.FolderIDs, f.ID)
			}

		case *tg.DialogFilterChatlist:
			// Папка-ссылка: состав задан явно, предикатов у неё нет.
			folders = append(folders, domain.Folder{
				ID:       f.ID,
				Title:    f.Title.Text,
				Emoticon: f.Emoticon,
				Order:    len(folders),
			})

			include := peerKeySet(ownID, f.PinnedPeers, f.IncludePeers)
			for i := range dialogs {
				d := &dialogs[i]
				if _, included := include[d.key]; included {
					d.chat.FolderIDs = append(d.chat.FolderIDs, f.ID)
				}
			}

		default:
			// DialogFilterDefault — это «Все чаты», отдельной вкладкой не приходит.
			continue
		}
	}

	return folders
}

func matchesPredicates(f *tg.DialogFilter, d *dialogInfo) bool {
	switch d.chat.Type {
	case domain.ChatTypeDirect:
		switch {
		case d.isBot:
			return f.Bots
		case d.isContact:
			return f.Contacts
		default:
			return f.NonContacts
		}
	case domain.ChatTypeGroup:
		return f.Groups
	case domain.ChatTypeChannel:
		return f.Broadcasts
	default:
		return false
	}
}

func peerKeySet(ownID int64, groups ...[]tg.InputPeerClass) map[string]struct{} {
	set := make(map[string]struct{})
	for _, group := range groups {
		for _, p := range group {
			if key, ok := inputPeerClassKey(p, ownID); ok {
				set[key] = struct{}{}
			}
		}
	}
	return set
}
