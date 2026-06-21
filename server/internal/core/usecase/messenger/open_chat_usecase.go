package messenger

import (
	"context"
	"telegleb/internal/core/usecase/message"
)

type OpenChatUseCase struct {
	messageRepo message.MessageRepository
}

func NewOpenChatUseCase(messageRepo message.MessageRepository) *OpenChatUseCase {
	return &OpenChatUseCase{
		messageRepo: messageRepo,
	}
}

func (uc *OpenChatUseCase) Execute(ctx context.Context, input OpenChatInput) (OpenChatOutput, error) {
	if err := input.Validate(); err != nil {
		return OpenChatOutput{}, err
	}
	
	messages, err := uc.messageRepo.GetChatHistory(ctx, input.SessionToken, input.ChatID, input.Limit, input.Offset)
	if err != nil {
		return OpenChatOutput{}, err
	}

	return OpenChatOutput{
		Messages: messages,
	}, nil
}
