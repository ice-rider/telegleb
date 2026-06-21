package messenger

import (
	"context"
	"telegleb/internal/core/usecase/message"
)

type SendMessageUseCase struct {
	messageRepo message.MessageRepository
}

func NewSendMessageUseCase(messageRepo message.MessageRepository) *SendMessageUseCase {
	return &SendMessageUseCase{
		messageRepo: messageRepo,
	}
}

func (uc *SendMessageUseCase) Execute(ctx context.Context, input SendMessageInput) (SendMessageOutput, error) {
	if err := input.Validate(); err != nil {
		return SendMessageOutput{}, err
	}

	msg, err := uc.messageRepo.SendMessage(ctx, input.SessionToken, input.ChatID, input.Content)
	if err != nil {
		return SendMessageOutput{}, err
	}

	return SendMessageOutput{
		Message: msg,
	}, nil
}