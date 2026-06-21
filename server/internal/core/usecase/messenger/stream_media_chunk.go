package messenger

import (
	"context"
	"telegleb/internal/core/usecase/media"
)

type StreamMediaChunkUseCase struct {
	mediaRepo media.MediaRepository
}

func NewStreamMediaChunkUseCase(mediaRepo media.MediaRepository) *StreamMediaChunkUseCase {
	return &StreamMediaChunkUseCase{
		mediaRepo: mediaRepo,
	}
}

func (uc *StreamMediaChunkUseCase) Execute(ctx context.Context, input StreamMediaChunkInput) (StreamMediaChunkOutput, error) {
	if err := input.Validate(); err != nil {
		return StreamMediaChunkOutput{}, err
	}

	chunk, err := uc.mediaRepo.DownloadMediaChunk(ctx, input.SessionToken, input.MediaID, input.Offset, input.Limit)
	if err != nil {
		return StreamMediaChunkOutput{}, ErrStreamingFailed
	}

	return StreamMediaChunkOutput{
		Chunk: chunk,
	}, nil
}