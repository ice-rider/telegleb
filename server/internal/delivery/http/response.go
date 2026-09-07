package http

import (
	"encoding/json"
	"log/slog"

	"github.com/valyala/fasthttp"
)

func writeJSON(ctx *fasthttp.RequestCtx, statusCode int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json; charset=utf-8")
		ctx.SetBodyString(`{"error":{"code":"INTERNAL","message":"failed to encode response"}}`)
		return
	}
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetBody(body)
}

// writeError сериализует тело, а не склеивает строки: кавычка или перевод
// строки внутри сообщения иначе ломает JSON.
func (s *Server) writeError(ctx *fasthttp.RequestCtx, err error) {
	status, apiErr := mapError(err)
	if status >= fasthttp.StatusInternalServerError {
		s.log.Error("request failed",
			slog.String("path", string(ctx.Path())),
			slog.String("code", apiErr.Code),
			slog.String("error", err.Error()),
		)
	} else {
		s.log.Warn("request rejected",
			slog.String("path", string(ctx.Path())),
			slog.String("code", apiErr.Code),
			slog.String("error", err.Error()),
		)
	}
	writeJSON(ctx, status, errorEnvelope{Error: apiErr})
}

func writeErrorCode(ctx *fasthttp.RequestCtx, status int, code, message string) {
	writeJSON(ctx, status, errorEnvelope{Error: apiError{Code: code, Message: message}})
}

func parseBody(ctx *fasthttp.RequestCtx, dst any) error {
	return json.Unmarshal(ctx.PostBody(), dst)
}
