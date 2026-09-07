package telegram

import (
	"context"
	"fmt"

	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/media"

	"github.com/gotd/td/tg"
)

type TelegramMediaRepository struct {
	adapter *TelegramAdapter
}

func NewTelegramMediaRepository(adapter *TelegramAdapter) media.MediaRepository {
	return &TelegramMediaRepository{adapter: adapter}
}

func (r *TelegramMediaRepository) DownloadChunk(ctx context.Context, sessionToken string, ref domain.MediaRef, offset int64, limit int) (media.Chunk, error) {
	client, err := r.adapter.GetClient(ctx, sessionToken)
	if err != nil {
		return media.Chunk{}, err
	}

	// fileReference протухает, поэтому его нельзя хранить в идентификаторе —
	// сообщение перезапрашивается, и reference берётся свежий.
	msg, err := fetchMessage(ctx, client.API, ref)
	if err != nil {
		return media.Chunk{}, err
	}

	location, mimeType, total, err := fileLocation(msg)
	if err != nil {
		return media.Chunk{}, err
	}

	resp, err := client.API.UploadGetFile(ctx, &tg.UploadGetFileRequest{
		Location: location,
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		return media.Chunk{}, fmt.Errorf("failed to download media chunk: %w", err)
	}

	file, ok := resp.(*tg.UploadFile)
	if !ok {
		return media.Chunk{}, fmt.Errorf("unexpected file response type from telegram api")
	}

	return media.Chunk{Bytes: file.Bytes, MimeType: mimeType, Total: total}, nil
}

func fetchMessage(ctx context.Context, api *tg.Client, ref domain.MediaRef) (*tg.Message, error) {
	ids := []tg.InputMessageClass{&tg.InputMessageID{ID: ref.MessageID}}

	var (
		resp tg.MessagesMessagesClass
		err  error
	)
	if ref.Peer.Kind == domain.PeerChannel {
		resp, err = api.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
			Channel: inputChannel(ref.Peer),
			ID:      ids,
		})
	} else {
		resp, err = api.MessagesGetMessages(ctx, ids)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to refetch message for media: %w", err)
	}

	msgs, _, _ := unpackMessages(resp)
	for _, m := range msgs {
		if msg, ok := m.(*tg.Message); ok && msg.ID == ref.MessageID {
			return msg, nil
		}
	}
	return nil, fmt.Errorf("message %d not found", ref.MessageID)
}

func fileLocation(msg *tg.Message) (tg.InputFileLocationClass, string, int64, error) {
	if msg.Media == nil {
		return nil, "", 0, fmt.Errorf("message has no media")
	}

	switch m := msg.Media.(type) {
	case *tg.MessageMediaDocument:
		doc, ok := m.Document.(*tg.Document)
		if !ok {
			return nil, "", 0, fmt.Errorf("document is unavailable")
		}
		return &tg.InputDocumentFileLocation{
			ID:            doc.ID,
			AccessHash:    doc.AccessHash,
			FileReference: doc.FileReference,
		}, doc.MimeType, doc.Size, nil

	case *tg.MessageMediaPhoto:
		photo, ok := m.Photo.(*tg.Photo)
		if !ok {
			return nil, "", 0, fmt.Errorf("photo is unavailable")
		}
		thumb, size := largestPhotoSize(photo)
		if thumb == "" {
			return nil, "", 0, fmt.Errorf("photo has no downloadable size")
		}
		return &tg.InputPhotoFileLocation{
			ID:            photo.ID,
			AccessHash:    photo.AccessHash,
			FileReference: photo.FileReference,
			ThumbSize:     thumb,
		}, "image/jpeg", size, nil

	default:
		return nil, "", 0, fmt.Errorf("unsupported media type")
	}
}

func largestPhotoSize(photo *tg.Photo) (string, int64) {
	var (
		bestType string
		bestArea int
		bestSize int64
	)
	for _, s := range photo.Sizes {
		size, ok := s.(*tg.PhotoSize)
		if !ok {
			continue
		}
		if area := size.W * size.H; area > bestArea {
			bestArea, bestType, bestSize = area, size.Type, int64(size.Size)
		}
	}
	return bestType, bestSize
}
