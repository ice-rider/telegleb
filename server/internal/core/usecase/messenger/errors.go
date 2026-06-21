package messenger

import "errors"

var (
	ErrChatNotFound        = errors.New("chat not found")
	ErrMediaNotFound       = errors.New("media file not found")
	ErrDashboardFailed     = errors.New("failed to load dashboard data")
	ErrStreamingFailed     = errors.New("failed to stream media chunk")
	ErrInvalidSessionToken = errors.New("invalid session token")
	ErrInvalidPagination   = errors.New("pagination parameters cannot be negative")
	ErrMediaOffset         = errors.New("media offset cannot be negative")
	ErrMediaChunkLimit     = errors.New("media chunk limit must be greater than zero")
	ErrEmptyMessage        = errors.New("message content cannot be empty")
)
