package telegram

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/chat"

	"github.com/gotd/td/tg"
	"golang.org/x/sync/errgroup"
)

const (
	// Telegram отдаёт максимум 100 диалогов за запрос.
	dialogPageSize = 100
	// Потолки на полную выборку. Список диалогов забирается ЦЕЛИКОМ, а не
	// первой страницей: принадлежность чата к папке считается по всему
	// списку, и если чат не попал в выборку, папка молча теряет участника.
	// Именно поэтому в папке из 13 чатов могло оказаться 2 — остальные
	// просто не входили в первую сотню по свежести.
	maxMainDialogs    = 1500
	maxArchiveDialogs = 500
)

type TelegramChatRepository struct {
	adapter *TelegramAdapter
}

func NewTelegramChatRepository(adapter *TelegramAdapter) chat.ChatRepository {
	return &TelegramChatRepository{adapter: adapter}
}

// dialogCursor — курсор пагинации диалогов. В MTProto это тройка
// (offset_date, offset_id, offset_peer), а не порядковый номер.
type dialogCursor struct {
	Date int    `json:"d"`
	ID   int    `json:"i"`
	Peer string `json:"p"`
}

func encodeCursor(c dialogCursor) string {
	raw, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeCursor(s string) (dialogCursor, bool) {
	if s == "" {
		return dialogCursor{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return dialogCursor{}, false
	}
	var c dialogCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return dialogCursor{}, false
	}
	return c, true
}

// dialogInfo — чат вместе с признаками, нужными только для раскладки по
// папкам. В домен они не попадают: это деталь предикатов Telegram.
type dialogInfo struct {
	chat      domain.Chat
	key       string
	isContact bool
	isBot     bool
}

type dialogPage struct {
	dialogs []dialogInfo
	next    string
	stats   domain.DashboardStats
}

func (r *TelegramChatRepository) GetDashboard(ctx context.Context, sessionToken string, limit int, cursor string) (domain.Dashboard, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return domain.Dashboard{}, err
	}
	api := client.API
	log := r.adapter.log

	// Профиль нужен раньше остального: без собственного id не отличить
	// «Избранное» от обычного диалога и не разрешить InputPeerSelf в папках.
	me, err := fetchMe(ctx, api)
	if err != nil {
		return domain.Dashboard{}, err
	}

	var (
		main    dialogPage
		archive dialogPage
		filters []tg.DialogFilterClass
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		main, err = fetchAllDialogs(gCtx, api, nil, maxMainDialogs, me.ID)
		return err
	})
	g.Go(func() error {
		archiveFolder := 1
		var err error
		archive, err = fetchAllDialogs(gCtx, api, &archiveFolder, maxArchiveDialogs, me.ID)
		return err
	})
	g.Go(func() error {
		res, err := api.MessagesGetDialogFilters(gCtx)
		if err != nil {
			return fmt.Errorf("failed to fetch dialog filters: %w", err)
		}
		filters = res.Filters
		return nil
	})
	if err := g.Wait(); err != nil {
		return domain.Dashboard{}, err
	}

	dialogs := mergeDialogs(main.dialogs, archive.dialogs)
	folders := assignFolders(dialogs, filters, me.ID)

	chats := make([]domain.Chat, 0, len(dialogs))
	var mainOrder, archiveOrder, archivedCount int
	for i := range dialogs {
		c := dialogs[i].chat
		if c.Archived {
			c.Order = archiveOrder
			archiveOrder++
			archivedCount++
		} else {
			c.Order = mainOrder
			mainOrder++
		}
		chats = append(chats, c)
	}

	stats := domain.DashboardStats{
		Pages:              main.stats.Pages + archive.stats.Pages,
		RawDialogs:         main.stats.RawDialogs + archive.stats.RawDialogs,
		SkippedUnknownPeer: main.stats.SkippedUnknownPeer + archive.stats.SkippedUnknownPeer,
		SkippedNotDialog:   main.stats.SkippedNotDialog + archive.stats.SkippedNotDialog,
	}
	truncated := main.next != "" || archive.next != ""

	// Сводка по загрузке. Она отвечает на вопрос «почему в папке мало чатов»
	// числами: сколько диалогов пришло, сколько отброшено и почему, и
	// сколько участников получила каждая папка.
	perFolder := make([]any, 0, len(folders)*2)
	counts := map[int]int{}
	for i := range dialogs {
		for _, fid := range dialogs[i].chat.FolderIDs {
			counts[fid]++
		}
	}
	for _, f := range folders {
		perFolder = append(perFolder, fmt.Sprintf("%s(%d)", f.Title, f.ID), counts[f.ID])
	}

	log.Info("dashboard loaded",
		slog.String("token", shortToken(sessionToken)),
		slog.Int("chats", len(chats)),
		slog.Int("archived", archivedCount),
		slog.Int("folders", len(folders)),
		slog.Int("pages", stats.Pages),
		slog.Int("raw_dialogs", stats.RawDialogs),
		slog.Int("skipped_unknown_peer", stats.SkippedUnknownPeer),
		slog.Int("skipped_not_dialog", stats.SkippedNotDialog),
		slog.Bool("truncated", truncated),
		slog.Group("per_folder", perFolder...),
	)

	if stats.SkippedUnknownPeer > 0 {
		log.Warn("часть диалогов отброшена: пир не найден в словарях ответа",
			slog.Int("count", stats.SkippedUnknownPeer))
	}
	if truncated {
		log.Warn("список диалогов упёрся в потолок — раскладка по папкам может быть неполной",
			slog.Int("limit_main", maxMainDialogs), slog.Int("limit_archive", maxArchiveDialogs))
	}

	return domain.Dashboard{
		Me:        me,
		Chats:     chats,
		Folders:   folders,
		Truncated: truncated,
		Stats:     stats,
	}, nil
}

func fetchMe(ctx context.Context, api *tg.Client) (domain.Me, error) {
	users, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
	if err != nil {
		return domain.Me{}, fmt.Errorf("failed to get own user: %w", err)
	}
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			return domain.Me{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Username:  user.Username,
				Phone:     user.Phone,
			}, nil
		}
	}
	return domain.Me{}, fmt.Errorf("own user not found in response")
}

// fetchAllDialogs выбирает диалоги страницами до исчерпания или до потолка.
// Возвращает непустой next, если упёрлись в потолок, — значит выборка неполна.
func fetchAllDialogs(ctx context.Context, api *tg.Client, folderID *int, max int, ownID int64) (dialogPage, error) {
	var (
		all    []dialogInfo
		stats  domain.DashboardStats
		cursor string
	)

	for len(all) < max {
		page, err := fetchDialogs(ctx, api, folderID, dialogPageSize, cursor, ownID)
		if err != nil {
			return dialogPage{}, err
		}

		all = append(all, page.dialogs...)
		stats.Pages++
		stats.RawDialogs += page.stats.RawDialogs
		stats.SkippedUnknownPeer += page.stats.SkippedUnknownPeer
		stats.SkippedNotDialog += page.stats.SkippedNotDialog

		if page.next == "" {
			return dialogPage{dialogs: all, stats: stats}, nil
		}
		cursor = page.next
	}

	return dialogPage{dialogs: all, next: cursor, stats: stats}, nil
}

// fetchDialogs забирает одну страницу. folderID == nil — запрос без фильтра
// по папке (основной список), иначе конкретная папка (1 — архив).
func fetchDialogs(ctx context.Context, api *tg.Client, folderID *int, limit int, cursor string, ownID int64) (dialogPage, error) {
	req := &tg.MessagesGetDialogsRequest{
		Limit:      limit,
		OffsetPeer: &tg.InputPeerEmpty{},
	}
	if folderID != nil {
		req.SetFolderID(*folderID)
	}
	if c, ok := decodeCursor(cursor); ok {
		req.OffsetDate = c.Date
		req.OffsetID = c.ID
		if p, err := domain.ParsePeer(c.Peer); err == nil {
			req.OffsetPeer = inputPeer(p)
		}
	}

	resp, err := api.MessagesGetDialogs(ctx, req)
	if err != nil {
		return dialogPage{}, fmt.Errorf("failed to fetch telegram dialogs: %w", err)
	}

	switch res := resp.(type) {
	case *tg.MessagesDialogsSlice:
		return buildPage(res.Dialogs, res.Chats, res.Users, res.Messages, ownID, len(res.Dialogs) >= limit), nil
	case *tg.MessagesDialogs:
		return buildPage(res.Dialogs, res.Chats, res.Users, res.Messages, ownID, false), nil
	default:
		return dialogPage{}, nil
	}
}

func buildPage(dialogs []tg.DialogClass, chats []tg.ChatClass, users []tg.UserClass, messages []tg.MessageClass, ownID int64, hasMore bool) dialogPage {
	d := newDicts(chats, users, ownID)
	d.addMessages(messages)

	now := int(time.Now().Unix())
	page := dialogPage{dialogs: make([]dialogInfo, 0, len(dialogs))}
	page.stats.RawDialogs = len(dialogs)

	var lastDialog *tg.Dialog
	for _, dClass := range dialogs {
		dlg, ok := dClass.(*tg.Dialog)
		if !ok {
			page.stats.SkippedNotDialog++
			continue
		}
		lastDialog = dlg

		info, ok := d.peerOf(dlg.Peer)
		if !ok {
			// Пир не пришёл в словарях ответа — восстановить accessHash
			// неоткуда. Считаем такие случаи, чтобы потеря была видна.
			page.stats.SkippedUnknownPeer++
			continue
		}

		folderID, _ := dlg.GetFolderID()
		muteUntil, _ := dlg.NotifySettings.GetMuteUntil()

		entry := dialogInfo{
			chat: domain.Chat{
				Peer:                info.Peer,
				Type:                info.Type,
				Title:               info.Title,
				IsForum:             info.IsForum,
				UnreadCount:         dlg.UnreadCount,
				UnreadMentionsCount: dlg.UnreadMentionsCount,
				MarkedUnread:        dlg.UnreadMark,
				Pinned:              dlg.Pinned,
				Muted:               muteUntil > now,
				Archived:            folderID == 1,
				FolderIDs:           []int{},
			},
			key: peerKey(info.Peer.Kind, info.Peer.ID),
		}

		if user, ok := d.users[info.Peer.ID]; ok && info.Peer.Kind == domain.PeerUser {
			entry.isContact = user.Contact
			entry.isBot = user.Bot
		}

		if topMsg, found := d.msgs[msgKey{peer: entry.key, msgID: dlg.TopMessage}]; found {
			msg := d.mapMessage(topMsg, info.Peer)
			entry.chat.LastMessage = &msg
		}

		page.dialogs = append(page.dialogs, entry)
	}

	if hasMore && lastDialog != nil {
		if info, ok := d.peerOf(lastDialog.Peer); ok {
			c := dialogCursor{ID: lastDialog.TopMessage, Peer: info.Peer.Ref()}
			if last, found := d.msgs[msgKey{peer: peerKey(info.Peer.Kind, info.Peer.ID), msgID: lastDialog.TopMessage}]; found {
				c.Date = last.Date
			}
			page.next = encodeCursor(c)
		}
	}

	return page
}

// mergeDialogs склеивает основной список с архивом, отбрасывая повторы.
// Запрос без folder_id на части слоёв уже возвращает архивные диалоги, а
// отдельный запрос архива гарантирует, что вкладка не будет пустой;
// дедупликация делает пересечение этих двух источников безобидным.
func mergeDialogs(main, archive []dialogInfo) []dialogInfo {
	seen := make(map[string]struct{}, len(main)+len(archive))
	merged := make([]dialogInfo, 0, len(main)+len(archive))

	for _, group := range [][]dialogInfo{main, archive} {
		for _, info := range group {
			if _, dup := seen[info.key]; dup {
				continue
			}
			seen[info.key] = struct{}{}
			merged = append(merged, info)
		}
	}
	return merged
}
