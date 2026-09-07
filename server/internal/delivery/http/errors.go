package http

import (
	"errors"

	"telegleb/internal/adapter/repository/session"
	"telegleb/internal/adapter/telegram"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/auth"
	"telegleb/internal/core/usecase/messenger"

	"github.com/valyala/fasthttp"
)

// apiError — единственная форма тела для любого ответа >= 400. Клиент
// ветвится по code; message предназначен человеку и контрактом не является.
type apiError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retryAfter,omitempty"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

// mapError переводит доменную ошибку в код ответа. Текст здесь всегда
// фиксированный: наружу не должны утекать обёрнутые ошибки gotd с внутренними
// деталями соединения.
func mapError(err error) (int, apiError) {
	if flood, ok := telegram.AsFloodWait(err); ok {
		return fasthttp.StatusTooManyRequests, apiError{
			Code:       "FLOOD_WAIT",
			Message:    "too many requests, retry later",
			RetryAfter: flood.RetryAfter,
		}
	}

	switch {
	case errors.Is(err, auth.ErrInvalidPhone):
		return fasthttp.StatusBadRequest, apiError{Code: "INVALID_PHONE", Message: "invalid phone number"}
	case errors.Is(err, auth.ErrEmptyInput), errors.Is(err, messenger.ErrInvalidRange):
		return fasthttp.StatusBadRequest, apiError{Code: "INVALID_BODY", Message: "invalid request body"}
	case errors.Is(err, auth.ErrInvalidCode), errors.Is(err, telegram.ErrTelegramSignIn):
		return fasthttp.StatusUnauthorized, apiError{Code: "INVALID_CODE", Message: "invalid confirmation code"}
	case errors.Is(err, auth.ErrInvalidPassword), errors.Is(err, telegram.ErrInvalid2FAPassword):
		return fasthttp.StatusUnauthorized, apiError{Code: "INVALID_PASSWORD", Message: "invalid password"}
	case errors.Is(err, auth.ErrSessionNotFound),
		errors.Is(err, telegram.ErrSessionNotFound),
		errors.Is(err, session.ErrSessionNotFound):
		return fasthttp.StatusUnauthorized, apiError{Code: "SESSION_EXPIRED", Message: "session expired"}
	case errors.Is(err, auth.ErrInvalidSessionState):
		return fasthttp.StatusConflict, apiError{Code: "SESSION_INVALID_STATE", Message: "authorization step out of order"}
	case errors.Is(err, messenger.ErrInvalidPeer), errors.Is(err, domain.ErrInvalidPeerRef):
		return fasthttp.StatusBadRequest, apiError{Code: "PEER_INVALID", Message: "malformed chat reference"}
	case errors.Is(err, messenger.ErrPeerNotFound):
		return fasthttp.StatusNotFound, apiError{Code: "PEER_NOT_FOUND", Message: "chat not found"}
	case errors.Is(err, messenger.ErrEmptyMessage):
		return fasthttp.StatusBadRequest, apiError{Code: "MESSAGE_EMPTY", Message: "message text is empty"}
	case errors.Is(err, messenger.ErrInvalidMediaRef),
		errors.Is(err, messenger.ErrMediaNotFound),
		errors.Is(err, domain.ErrInvalidMediaRef):
		return fasthttp.StatusNotFound, apiError{Code: "MEDIA_NOT_FOUND", Message: "media not found"}
	case errors.Is(err, telegram.ErrClientInitFailed), errors.Is(err, telegram.ErrTelegramSendCode):
		return fasthttp.StatusBadGateway, apiError{Code: "TELEGRAM_ERROR", Message: "telegram is unavailable"}
	default:
		return fasthttp.StatusInternalServerError, apiError{Code: "INTERNAL", Message: "internal server error"}
	}
}
