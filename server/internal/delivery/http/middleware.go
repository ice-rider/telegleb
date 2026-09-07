package http

import (
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

const bearerPrefix = "Bearer "

// bearerToken достаёт токен сессии из заголовка и проверяет его подпись.
// Тело запроса и query-строка для этого не используются: токен в query
// оседает в access-логах.
//
// Подпись проверяется до похода в Redis — иначе токен работал бы просто как
// ключ хранилища, и подошла бы любая строка, которая туда попала.
func (s *Server) bearerToken(ctx *fasthttp.RequestCtx) (string, bool) {
	header := string(ctx.Request.Header.Peek("Authorization"))
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(bearerPrefix):])
	if token == "" {
		return "", false
	}

	if _, err := s.tokens.ValidateToken(token); err != nil {
		s.log.Warn("rejected token with invalid signature", slog.String("path", string(ctx.Path())))
		return "", false
	}

	return token, true
}

func middlewareRecover(log *slog.Logger, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					slog.Any("error", r),
					slog.String("path", string(ctx.Path())),
					slog.String("stack", string(debug.Stack())),
				)
				writeErrorCode(ctx, fasthttp.StatusInternalServerError, "INTERNAL", "internal server error")
			}
		}()
		next(ctx)
	}
}

// middlewareLogger пишет запись после обработки, иначе статус всегда 200.
func middlewareLogger(log *slog.Logger, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		next(ctx)
		log.Debug("request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.Int("status", ctx.Response.StatusCode()),
			slog.Duration("elapsed", time.Since(start)),
		)
	}
}
