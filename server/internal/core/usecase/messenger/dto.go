package messenger

import (
	"strings"
	"telegleb/internal/core/domain"
)

type LoadDashboardInput struct {
	SessionToken string
	Limit        int
	Offset       int
}

func (i LoadDashboardInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionToken
	}
	if i.Limit < 0 || i.Offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

type LoadDashboardOutput struct {
	Chats      []domain.Chat
	Folders    []domain.Folder
	OwnUserID  int64
}

type OpenChatInput struct {
	SessionToken string
	ChatID       string
	Limit        int
	Offset       int
}

func (i OpenChatInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionToken
	}
	if strings.TrimSpace(i.ChatID) == "" {
		return ErrChatNotFound
	}
	if i.Limit < 0 || i.Offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

type OpenChatOutput struct {
	Messages []domain.Message
}

type SendMessageInput struct {
	SessionToken string
	ChatID       string
	Content      string
}

func (i SendMessageInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionToken
	}
	if strings.TrimSpace(i.ChatID) == "" {
		return ErrChatNotFound
	}
	if strings.TrimSpace(i.Content) == "" {
		return ErrEmptyMessage
	}
	return nil
}

type SendMessageOutput struct {
	Message domain.Message
}

type DownloadMediaLinkInput struct {
	SessionToken string
	MediaID      string
}

func (i DownloadMediaLinkInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionToken
	}
	if strings.TrimSpace(i.MediaID) == "" {
		return ErrMediaNotFound
	}
	return nil
}

type DownloadMediaLinkOutput struct {
	Link string
}

type StreamMediaChunkInput struct {
	SessionToken string
	MediaID      string
	Offset       int64
	Limit        int
}

func (i StreamMediaChunkInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionToken
	}
	if strings.TrimSpace(i.MediaID) == "" {
		return ErrMediaNotFound
	}
	if i.Offset < 0 {
		return ErrMediaOffset
	}
	if i.Limit <= 0 {
		return ErrMediaChunkLimit
	}
	return nil
}

type StreamMediaChunkOutput struct {
	Chunk []byte
}
