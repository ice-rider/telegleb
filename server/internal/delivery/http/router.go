package http

import (
	"bytes"

	"github.com/valyala/fasthttp"
)

func (s *Server) setupRouter() fasthttp.RequestHandler {
	authPrefix := []byte("/api/v1/auth/")
	messengerPrefix := []byte("/api/v1/messenger/")
	mediaPrefix := []byte("/api/v1/media/stream/")
	methodGet := []byte("GET")
	methodOptions := []byte("OPTIONS")

	return func(ctx *fasthttp.RequestCtx) {
		path := ctx.Path()
		method := ctx.Method()

		if bytes.Equal(method, methodOptions) {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		if bytes.HasPrefix(path, authPrefix) {
			s.handleAuthRoutes(ctx, path[len(authPrefix):], method)
			return
		}

		if bytes.HasPrefix(path, messengerPrefix) {
			s.handleMessengerRoutes(ctx, path[len(messengerPrefix):], method)
			return
		}

		if bytes.HasPrefix(path, mediaPrefix) && bytes.Equal(method, methodGet) {
			s.handleStreamMedia(ctx)
			return
		}

		writeError(ctx, fasthttp.StatusNotFound, "not found")
	}
}

func (s *Server) handleAuthRoutes(ctx *fasthttp.RequestCtx, action []byte, method []byte) {
	if !bytes.Equal(method, []byte("POST")) {
		writeError(ctx, fasthttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch {
	case bytes.Equal(action, []byte("request-login")):
		s.handleRequestLogin(ctx)
	case bytes.Equal(action, []byte("verify-code")):
		s.handleVerifyCode(ctx)
	case bytes.Equal(action, []byte("verify-password")):
		s.handleVerifyPassword(ctx)
	case bytes.Equal(action, []byte("logout")):
		s.handleLogout(ctx)
	default:
		writeError(ctx, fasthttp.StatusNotFound, "not found")
	}
}

func (s *Server) handleMessengerRoutes(ctx *fasthttp.RequestCtx, action []byte, method []byte) {
	if !bytes.Equal(method, []byte("POST")) {
		writeError(ctx, fasthttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch {
	case bytes.Equal(action, []byte("dashboard")):
		s.handleLoadDashboard(ctx)
	case bytes.Equal(action, []byte("open-chat")):
		s.handleOpenChat(ctx)
	case bytes.Equal(action, []byte("send-message")):
		s.handleSendMessage(ctx)
	default:
		writeError(ctx, fasthttp.StatusNotFound, "not found")
	}
}
