package messenger

import (
	"context"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/chat"

	"golang.org/x/sync/errgroup"
)

type LoadDashboardUseCase struct {
	chatRepo chat.ChatRepository
}

func NewLoadDashboardUseCase(chatRepo chat.ChatRepository) *LoadDashboardUseCase {
	return &LoadDashboardUseCase{
		chatRepo: chatRepo,
	}
}

func (uc *LoadDashboardUseCase) Execute(ctx context.Context, input LoadDashboardInput) (LoadDashboardOutput, error) {
	if err := input.Validate(); err != nil {
		return LoadDashboardOutput{}, err
	}

	g, gCtx := errgroup.WithContext(ctx)

	var chats []domain.Chat
	var folders []domain.Folder
	var ownUserID int64

	g.Go(func() error {
		var err error
		chats, err = uc.chatRepo.GetChats(gCtx, input.SessionToken, input.Limit, input.Offset)
		return err
	})

	g.Go(func() error {
		var err error
		folders, err = uc.chatRepo.GetFolders(gCtx, input.SessionToken)
		return err
	})

	g.Go(func() error {
		var err error
		ownUserID, err = uc.chatRepo.GetOwnUserID(gCtx, input.SessionToken)
		return err
	})

	if err := g.Wait(); err != nil {
		return LoadDashboardOutput{}, ErrDashboardFailed
	}

	return LoadDashboardOutput{
		Chats:     chats,
		Folders:   folders,
		OwnUserID: ownUserID,
	}, nil
}