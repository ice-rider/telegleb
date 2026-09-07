package messenger

import (
	"context"

	"telegleb/internal/core/usecase/chat"
	"telegleb/internal/core/usecase/media"
	"telegleb/internal/core/usecase/message"
)

// LoadDashboardUseCase отдаёт чаты, папки и профиль одним снимком. Раздельная
// загрузка здесь недопустима: клиент не должен уметь отрисовать чат раньше,
// чем узнает его папку.
type LoadDashboardUseCase struct {
	chatRepo chat.ChatRepository
}

func NewLoadDashboardUseCase(chatRepo chat.ChatRepository) *LoadDashboardUseCase {
	return &LoadDashboardUseCase{chatRepo: chatRepo}
}

func (uc *LoadDashboardUseCase) Execute(ctx context.Context, input LoadDashboardInput) (LoadDashboardOutput, error) {
	dashboard, err := uc.chatRepo.GetDashboard(ctx, input.SessionToken, clampLimit(input.Limit), input.Cursor)
	if err != nil {
		return LoadDashboardOutput{}, err
	}
	return LoadDashboardOutput{Dashboard: dashboard}, nil
}

type OpenChatUseCase struct {
	messageRepo message.MessageRepository
}

func NewOpenChatUseCase(messageRepo message.MessageRepository) *OpenChatUseCase {
	return &OpenChatUseCase{messageRepo: messageRepo}
}

func (uc *OpenChatUseCase) Execute(ctx context.Context, input OpenChatInput) (OpenChatOutput, error) {
	peer, err := input.Peer()
	if err != nil {
		return OpenChatOutput{}, err
	}

	limit := clampLimit(input.Limit)
	messages, err := uc.messageRepo.GetHistory(ctx, input.SessionToken, peer, limit, input.BeforeID)
	if err != nil {
		return OpenChatOutput{}, err
	}

	// Сообщения отсортированы по возрастанию, значит самое старое — первое.
	// Оно и становится границей следующей страницы.
	var nextBeforeID int
	if len(messages) == limit && len(messages) > 0 {
		nextBeforeID = int(messages[0].ID)
	}

	return OpenChatOutput{Messages: messages, NextBeforeID: nextBeforeID}, nil
}

type SendMessageUseCase struct {
	messageRepo message.MessageRepository
}

func NewSendMessageUseCase(messageRepo message.MessageRepository) *SendMessageUseCase {
	return &SendMessageUseCase{messageRepo: messageRepo}
}

func (uc *SendMessageUseCase) Execute(ctx context.Context, input SendMessageInput) (SendMessageOutput, error) {
	if err := input.Validate(); err != nil {
		return SendMessageOutput{}, err
	}
	peer, err := input.Peer()
	if err != nil {
		return SendMessageOutput{}, err
	}

	msg, err := uc.messageRepo.Send(ctx, input.SessionToken, peer, input.Text, input.randomID())
	if err != nil {
		return SendMessageOutput{}, err
	}
	return SendMessageOutput{Message: msg}, nil
}

type StreamMediaUseCase struct {
	mediaRepo media.MediaRepository
}

func NewStreamMediaUseCase(mediaRepo media.MediaRepository) *StreamMediaUseCase {
	return &StreamMediaUseCase{mediaRepo: mediaRepo}
}

func (uc *StreamMediaUseCase) Execute(ctx context.Context, input StreamMediaInput) (media.Chunk, error) {
	if err := input.Validate(); err != nil {
		return media.Chunk{}, err
	}
	ref, err := input.Ref()
	if err != nil {
		return media.Chunk{}, err
	}
	return uc.mediaRepo.DownloadChunk(ctx, input.SessionToken, ref, input.Offset, input.Limit)
}
