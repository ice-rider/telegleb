package http

import (
	"log/slog"
	"runtime/debug"

	"github.com/valyala/fasthttp"
)

func middlewareCORS(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if ctx.IsOptions() {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		next(ctx)
	}
}

func middlewareRecover(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "error", r, "stack", string(debug.Stack()))
				writeError(ctx, fasthttp.StatusInternalServerError, "internal server error")
			}
		}()
		next(ctx)
	}
}

func middlewareLogger(log *slog.Logger, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		log.Info("request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.Int("status", ctx.Response.StatusCode()),
		)
		next(ctx)
	}
}
