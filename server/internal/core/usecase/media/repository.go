package media

import (
	"context"

	"telegleb/internal/core/domain"
)

type Chunk struct {
	Bytes    []byte
	MimeType string
	Total    int64
}

type MediaRepository interface {
	DownloadChunk(ctx context.Context, sessionToken string, ref domain.MediaRef, offset int64, limit int) (Chunk, error)
}
