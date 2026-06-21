package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
	"telegleb/internal/core/usecase/media"
)

type TelegramMediaRepository struct {
	adapter *TelegramAdapter
}

func NewTelegramMediaRepository(adapter *TelegramAdapter) media.MediaRepository {
	return &TelegramMediaRepository{adapter: adapter}
}

func (r *TelegramMediaRepository) DownloadMediaLink(ctx context.Context, sessionToken string, mediaID string) (string, error) {
	return fmt.Sprintf("/api/v1/media/stream/%s", mediaID), nil
}

func (r *TelegramMediaRepository) DownloadMediaChunk(ctx context.Context, sessionToken string, mediaID string, offset int64, limit int) ([]byte, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// Разбираем наш составной ID
	parts := strings.Split(mediaID, "_")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid media id format")
	}

	mediaType := parts[0]
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid media structural component id")
	}
	accessHash, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid media structural component hash")
	}

	var location tg.InputFileLocationClass

	if mediaType == "doc" {
		location = &tg.InputDocumentFileLocation{
			ID:            id,
			AccessHash:    accessHash,
			FileReference: []byte{}, // Для тестов пустой референс, в продакшене извлекается из сообщения
		}
	} else {
		location = &tg.InputPhotoFileLocation{
			ID:            id,
			AccessHash:    accessHash,
			FileReference: []byte{},
			ThumbSize:     "x", // Запрашиваем стандартный большой размер картинки
		}
	}

	// Делаем прямой запрос к дата-центрам Telegram для получения байтового сегмента
	resp, err := client.API.UploadGetFile(ctx, &tg.UploadGetFileRequest{
		Location: location,
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download media chunk from telegram: %w", err)
	}

	switch result := resp.(type) {
	case *tg.UploadFile:
		return result.Bytes, nil
	default:
		return nil, fmt.Errorf("unexpected file response type from telegram api")
	}
}