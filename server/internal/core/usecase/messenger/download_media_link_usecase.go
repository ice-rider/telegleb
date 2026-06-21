package messenger

import (
	"context"
	"telegleb/internal/core/usecase/media"
)

type DownloadMediaLinkUseCase struct {
	mediaRepo media.MediaRepository
}

func NewDownloadMediaLinkUseCase(mediaRepo media.MediaRepository) *DownloadMediaLinkUseCase {
	return &DownloadMediaLinkUseCase{
		mediaRepo: mediaRepo,
	}
}

func (uc *DownloadMediaLinkUseCase) Execute(ctx context.Context, input DownloadMediaLinkInput) (DownloadMediaLinkOutput, error) {
	if err := input.Validate(); err != nil {
		return DownloadMediaLinkOutput{}, err
	}

	link, err := uc.mediaRepo.DownloadMediaLink(ctx, input.SessionToken, input.MediaID)
	if err != nil {
		return DownloadMediaLinkOutput{}, ErrMediaNotFound
	}

	return DownloadMediaLinkOutput{
		Link: link,
	}, nil
}