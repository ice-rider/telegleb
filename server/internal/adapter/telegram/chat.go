package telegram

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/chat"

	"github.com/gotd/td/tg"
	"golang.org/x/sync/errgroup"
)

// archivePageLimit — сколько архивных диалогов забираем вместе с первой
// страницей. Архив нужен клиенту целиком сразу, чтобы вкладка «Архив» никогда
// не оказалась наполовину заполненной.
const archivePageLimit = 100

type TelegramChatRepository struct {
	adapter *TelegramAdapter
}

func NewTelegramChatRepository(adapter *TelegramAdapter) chat.ChatRepository {
	return &TelegramChatRepository{adapter: adapter}
}

// dialogCursor — курсор пагинации диалогов. В MTProto это тройка
// (offset_date, offset_id, offset_peer), а не порядковый номер: смещение
// числом здесь не работает в принципе.
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
}

func (r *TelegramChatRepository) GetDashboard(ctx context.Context, sessionToken string, limit int, cursor string) (domain.Dashboard, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return domain.Dashboard{}, err
	}
	api := client.API

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

	firstPage := cursor == ""

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		main, err = fetchDialogs(gCtx, api, nil, limit, cursor, me.ID)
		return err
	})
	if firstPage {
		archiveFolder := 1
		g.Go(func() error {
			var err error
			archive, err = fetchDialogs(gCtx, api, &archiveFolder, archivePageLimit, "", me.ID)
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
	}
	if err := g.Wait(); err != nil {
		return domain.Dashboard{}, err
	}

	dialogs := mergeDialogs(main.dialogs, archive.dialogs)
	folders := assignFolders(dialogs, filters, me.ID)

	chats := make([]domain.Chat, 0, len(dialogs))
	var mainOrder, archiveOrder int
	for i := range dialogs {
		c := dialogs[i].chat
		if c.Archived {
			c.Order = archiveOrder
			archiveOrder++
		} else {
			c.Order = mainOrder
			mainOrder++
		}
		chats = append(chats, c)
	}

	return domain.Dashboard{
		Me:         me,
		Chats:      chats,
		Folders:    folders,
		NextCursor: main.next,
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

// fetchDialogs забирает одну страницу диалогов. folderID == nil — запрос без
// фильтра по папке (основной список), иначе конкретная папка Telegram
// (0 — основная, 1 — архив).
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

	var lastDialog *tg.Dialog
	for _, dClass := range dialogs {
		dlg, ok := dClass.(*tg.Dialog)
		if !ok {
			continue
		}

		peer, chatType, title, ok := d.peerOf(dlg.Peer)
		if !ok {
			continue
		}

		folderID, _ := dlg.GetFolderID()
		muteUntil, _ := dlg.NotifySettings.GetMuteUntil()

		info := dialogInfo{
			chat: domain.Chat{
				Peer:                peer,
				Type:                chatType,
				Title:               title,
				UnreadCount:         dlg.UnreadCount,
				UnreadMentionsCount: dlg.UnreadMentionsCount,
				MarkedUnread:        dlg.UnreadMark,
				Pinned:              dlg.Pinned,
				Muted:               muteUntil > now,
				Archived:            folderID == 1,
				FolderIDs:           []int{},
			},
			key: peerKey(peer.Kind, peer.ID),
		}

		if user, ok := d.users[peer.ID]; ok && peer.Kind == domain.PeerUser {
			info.isContact = user.Contact
			info.isBot = user.Bot
		}

		if topMsg, found := d.msgs[msgKey{peer: info.key, msgID: dlg.TopMessage}]; found {
			msg := d.mapMessage(topMsg, peer)
			info.chat.LastMessage = &msg
		}

		page.dialogs = append(page.dialogs, info)
		lastDialog = dlg
	}

	if hasMore && lastDialog != nil {
		if peer, _, _, ok := d.peerOf(lastDialog.Peer); ok {
			c := dialogCursor{ID: lastDialog.TopMessage, Peer: peer.Ref()}
			if last, found := d.msgs[msgKey{peer: peerKey(peer.Kind, peer.ID), msgID: lastDialog.TopMessage}]; found {
				c.Date = last.Date
			}
			page.next = encodeCursor(c)
		}
	}

	return page
}

// mergeDialogs склеивает основной список с архивом, отбрасывая повторы.
// Запрос без folder_id на части слоёв уже возвращает архивные диалоги, а
// отдельный запрос архива гарантирует, что вкладка не будет пустой; дедупликация
// делает пересечение этих двух источников безобидным.
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
