package http

import (
	"strconv"

	"telegleb/internal/core/usecase/messenger"

	"github.com/valyala/fasthttp"
)

type chatsResponse struct {
	Chats      []chatDTO   `json:"chats"`
	Folders    []folderDTO `json:"folders"`
	Me         meDTO       `json:"me"`
	NextCursor string      `json:"nextCursor,omitempty"`
}

type historyResponse struct {
	Messages     []messageDTO `json:"messages"`
	NextBeforeID int          `json:"nextBeforeId,omitempty"`
}

type sendMessageRequest struct {
	Text     string `json:"text"`
	RandomID string `json:"randomId"`
}

type sendMessageResponse struct {
	Message messageDTO `json:"message"`
}

// handleListChats отдаёт чаты, папки и профиль одним ответом. Разделять их на
// отдельные запросы нельзя: между ними существует момент, когда чаты уже
// пришли, а их принадлежность к папкам — ещё нет.
func (s *Server) handleListChats(ctx *fasthttp.RequestCtx) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	limit, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("limit")))
	cursor := string(ctx.QueryArgs().Peek("cursor"))

	output, err := s.loadDashboardUC.Execute(ctx, messenger.LoadDashboardInput{
		SessionToken: token,
		Limit:        limit,
		Cursor:       cursor,
	})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	dashboard := output.Dashboard
	chats := make([]chatDTO, 0, len(dashboard.Chats))
	for _, c := range dashboard.Chats {
		chats = append(chats, mapChat(c))
	}
	folders := make([]folderDTO, 0, len(dashboard.Folders))
	for _, f := range dashboard.Folders {
		folders = append(folders, mapFolder(f))
	}

	writeJSON(ctx, fasthttp.StatusOK, chatsResponse{
		Chats:      chats,
		Folders:    folders,
		Me:         mapMe(dashboard.Me),
		NextCursor: dashboard.NextCursor,
	})
}

func (s *Server) handleChatHistory(ctx *fasthttp.RequestCtx, chatRef string) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	limit, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("limit")))
	beforeID, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("beforeId")))

	output, err := s.openChatUC.Execute(ctx, messenger.OpenChatInput{
		SessionToken: token,
		ChatRef:      chatRef,
		Limit:        limit,
		BeforeID:     beforeID,
	})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	messages := make([]messageDTO, 0, len(output.Messages))
	for _, m := range output.Messages {
		messages = append(messages, mapMessage(m))
	}

	writeJSON(ctx, fasthttp.StatusOK, historyResponse{
		Messages:     messages,
		NextBeforeID: output.NextBeforeID,
	})
}

func (s *Server) handleSendMessage(ctx *fasthttp.RequestCtx, chatRef string) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	var req sendMessageRequest
	if err := parseBody(ctx, &req); err != nil {
		writeErrorCode(ctx, fasthttp.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	output, err := s.sendMessageUC.Execute(ctx, messenger.SendMessageInput{
		SessionToken: token,
		ChatRef:      chatRef,
		Text:         req.Text,
		RandomID:     req.RandomID,
	})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, sendMessageResponse{Message: mapMessage(output.Message)})
}

func (s *Server) handleHealth(ctx *fasthttp.RequestCtx) {
	status := map[string]string{"status": "ok", "redis": "ok"}
	code := fasthttp.StatusOK
	if err := s.redis.Ping(ctx).Err(); err != nil {
		status["status"] = "degraded"
		status["redis"] = "unavailable"
		code = fasthttp.StatusServiceUnavailable
	}
	writeJSON(ctx, code, status)
}
