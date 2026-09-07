package messenger

import "errors"

var (
	ErrInvalidPeer     = errors.New("chat reference is malformed")
	ErrPeerNotFound    = errors.New("chat not found or inaccessible")
	ErrEmptyMessage    = errors.New("message text cannot be empty")
	ErrInvalidMediaRef = errors.New("media reference is malformed")
	ErrMediaNotFound   = errors.New("media file not found")
	ErrInvalidRange    = errors.New("invalid offset or limit")
)
