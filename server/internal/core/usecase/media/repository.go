package media

import "context"

type MediaRepository interface {
	DownloadMediaLink(ctx context.Context, sessionToken string, mediaID string) (string, error)
	DownloadMediaChunk(ctx context.Context, sessionToken string, mediaID string, offset int64, limit int) ([]byte, error)
}