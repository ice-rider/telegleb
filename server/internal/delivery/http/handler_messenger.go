package http

import (
	"errors"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/messenger"

	"github.com/valyala/fasthttp"
)

type loadDashboardRequest struct {
	SessionToken string
	Limit        int
	Offset       int
}

type chatDTO struct {
	ID          int64      `json:"ID"`
	Title       string     `json:"Title"`
	Type        int        `json:"Type"`
	UnreadCount int        `json:"UnreadCount"`
	LastMessage messageDTO `json:"LastMessage"`
}

type messageEntityDTO struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
	URL    string `json:"url,omitempty"`
	UserID int64  `json:"userId,omitempty"`
}

type messageDTO struct {
	ID        int64  `json:"ID"`
	ChatID    int64  `json:"ChatID"`
	SenderID  int64  `json:"SenderID"`
	Text      string `json:"Text"`
	CreatedAt string `json:"CreatedAt"`
	HasMedia  bool   `json:"HasMedia"`
	MediaId   string `json:"MediaId"`

	Out        bool   `json:"Out"`
	Mentioned  bool   `json:"Mentioned"`
	Silent     bool   `json:"Silent"`
	Post       bool   `json:"Post"`
	Pinned     bool   `json:"Pinned"`
	Noforwards bool   `json:"Noforwards"`
	EditDate   string `json:"EditDate,omitempty"`
	Views      int    `json:"Views"`
	Forwards   int    `json:"Forwards"`
	GroupedID  int64  `json:"GroupedID"`
	ViaBotID   int64  `json:"ViaBotID"`
	PostAuthor string `json:"PostAuthor"`
	TTLPeriod  int    `json:"TTLPeriod"`

	ReplyToMsgID int64 `json:"ReplyToMsgID"`
	ReplyToPeer  int64 `json:"ReplyToPeer"`

	FwdFromName      string `json:"FwdFromName"`
	FwdFromDate      string `json:"FwdFromDate,omitempty"`
	FwdFromChannelID int64  `json:"FwdFromChannelID"`
	FwdFromUserID    int64  `json:"FwdFromUserID"`

	RepliesCount int   `json:"RepliesCount"`
	RepliesMaxID int64 `json:"RepliesMaxID"`

	Entities []messageEntityDTO `json:"Entities"`
}

type folderDTO struct {
	ID      int     `json:"ID"`
	Title   string  `json:"Title"`
	ChatIDs []int64 `json:"ChatIDs"`
}

type loadDashboardResponse struct {
	Chats      []chatDTO   `json:"chats"`
	Folders    []folderDTO `json:"folders"`
	OwnUserID  int64       `json:"OwnUserID"`
}

type openChatRequest struct {
	SessionToken string
	ChatID       string
	Limit        int
	Offset       int
}

type openChatResponse struct {
	Messages []messageDTO `json:"messages"`
}

type sendMessageRequest struct {
	SessionToken string
	ChatID       string
	Content      string
}

type sendMessageResponse struct {
	Message messageDTO `json:"message"`
}

func mapDomainChat(c domain.Chat) chatDTO {
	return chatDTO{
		ID:          c.ID,
		Title:       c.Title,
		Type:        int(c.Type),
		UnreadCount: c.UnreadCount,
		LastMessage: mapDomainMessage(c.LastMessage),
	}
}

func mapDomainMessage(m domain.Message) messageDTO {
	editDate := ""
	if !m.EditDate.IsZero() {
		editDate = m.EditDate.Format("2006-01-02T15:04:05Z")
	}
	fwdFromDate := ""
	if !m.FwdFromDate.IsZero() {
		fwdFromDate = m.FwdFromDate.Format("2006-01-02T15:04:05Z")
	}

	entities := make([]messageEntityDTO, 0, len(m.Entities))
	for _, e := range m.Entities {
		entities = append(entities, messageEntityDTO{
			Offset: e.Offset,
			Length: e.Length,
			Type:   e.Type,
			URL:    e.URL,
			UserID: e.UserID,
		})
	}

	return messageDTO{
		ID:        m.ID,
		ChatID:    m.ChatID,
		SenderID:  m.SenderID,
		Text:      m.Text,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		HasMedia:  m.HasMedia,
		MediaId:   m.MediaId,

		Out:        m.Out,
		Mentioned:  m.Mentioned,
		Silent:     m.Silent,
		Post:       m.Post,
		Pinned:     m.Pinned,
		Noforwards: m.Noforwards,
		EditDate:   editDate,
		Views:      m.Views,
		Forwards:   m.Forwards,
		GroupedID:  m.GroupedID,
		ViaBotID:   m.ViaBotID,
		PostAuthor: m.PostAuthor,
		TTLPeriod:  m.TTLPeriod,

		ReplyToMsgID: m.ReplyToMsgID,
		ReplyToPeer:  m.ReplyToPeer,

		FwdFromName:      m.FwdFromName,
		FwdFromDate:      fwdFromDate,
		FwdFromChannelID: m.FwdFromChannelID,
		FwdFromUserID:    m.FwdFromUserID,

		RepliesCount: m.RepliesCount,
		RepliesMaxID: m.RepliesMaxID,

		Entities: entities,
	}
}

func mapDomainFolder(f domain.Folder) folderDTO {
	return folderDTO{
		ID:      f.ID,
		Title:   f.Title,
		ChatIDs: f.ChatIDs,
	}
}

func mapMessengerErr(err error) int {
	switch {
	case errors.Is(err, messenger.ErrInvalidSessionToken),
		errors.Is(err, messenger.ErrChatNotFound):
		return fasthttp.StatusUnauthorized
	case errors.Is(err, messenger.ErrInvalidPagination),
		errors.Is(err, messenger.ErrEmptyMessage),
		errors.Is(err, messenger.ErrMediaOffset),
		errors.Is(err, messenger.ErrMediaChunkLimit):
		return fasthttp.StatusBadRequest
	case errors.Is(err, messenger.ErrMediaNotFound):
		return fasthttp.StatusNotFound
	default:
		return fasthttp.StatusInternalServerError
	}
}

func (s *Server) handleLoadDashboard(ctx *fasthttp.RequestCtx) {
	var req loadDashboardRequest
	if err := parseBody(ctx, &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	output, err := s.loadDashboardUC.Execute(ctx, messenger.LoadDashboardInput{
		SessionToken: req.SessionToken,
		Limit:        req.Limit,
		Offset:       req.Offset,
	})
	if err != nil {
		writeError(ctx, mapMessengerErr(err), err.Error())
		return
	}

	chats := make([]chatDTO, len(output.Chats))
	for i, c := range output.Chats {
		chats[i] = mapDomainChat(c)
	}
	folders := make([]folderDTO, len(output.Folders))
	for i, f := range output.Folders {
		folders[i] = mapDomainFolder(f)
	}

	writeJSON(ctx, fasthttp.StatusOK, loadDashboardResponse{
		Chats:     chats,
		Folders:   folders,
		OwnUserID: output.OwnUserID,
	})
}

func (s *Server) handleOpenChat(ctx *fasthttp.RequestCtx) {
	var req openChatRequest
	if err := parseBody(ctx, &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	output, err := s.openChatUC.Execute(ctx, messenger.OpenChatInput{
		SessionToken: req.SessionToken,
		ChatID:       req.ChatID,
		Limit:        req.Limit,
		Offset:       req.Offset,
	})
	if err != nil {
		writeError(ctx, mapMessengerErr(err), err.Error())
		return
	}

	msgs := make([]messageDTO, len(output.Messages))
	for i, m := range output.Messages {
		msgs[i] = mapDomainMessage(m)
	}

	writeJSON(ctx, fasthttp.StatusOK, openChatResponse{
		Messages: msgs,
	})
}

func (s *Server) handleSendMessage(ctx *fasthttp.RequestCtx) {
	var req sendMessageRequest
	if err := parseBody(ctx, &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	output, err := s.sendMessageUC.Execute(ctx, messenger.SendMessageInput{
		SessionToken: req.SessionToken,
		ChatID:       req.ChatID,
		Content:      req.Content,
	})
	if err != nil {
		writeError(ctx, mapMessengerErr(err), err.Error())
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, sendMessageResponse{
		Message: mapDomainMessage(output.Message),
	})
}
