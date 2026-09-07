package http

import (
	"strings"

	"github.com/valyala/fasthttp"
)

const apiPrefix = "api/v1"

func (s *Server) setupRouter() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := strings.Trim(string(ctx.Path()), "/")
		if !strings.HasPrefix(path, apiPrefix) {
			notFound(ctx)
			return
		}

		rest := strings.Trim(strings.TrimPrefix(path, apiPrefix), "/")
		if rest == "" {
			notFound(ctx)
			return
		}

		segments := strings.Split(rest, "/")
		method := string(ctx.Method())

		switch segments[0] {
		case "health":
			if len(segments) == 1 && method == fasthttp.MethodGet {
				s.handleHealth(ctx)
				return
			}
		case "auth":
			if len(segments) == 2 {
				s.routeAuth(ctx, segments[1], method)
				return
			}
		case "chats":
			switch {
			case len(segments) == 1 && method == fasthttp.MethodGet:
				s.handleListChats(ctx)
				return
			case len(segments) == 3 && segments[2] == "messages":
				switch method {
				case fasthttp.MethodGet:
					s.handleChatHistory(ctx, segments[1])
					return
				case fasthttp.MethodPost:
					s.handleSendMessage(ctx, segments[1])
					return
				}
			}
		case "media":
			if len(segments) == 2 && method == fasthttp.MethodGet {
				s.handleStreamMedia(ctx, segments[1])
				return
			}
		}

		notFound(ctx)
	}
}

func (s *Server) routeAuth(ctx *fasthttp.RequestCtx, action, method string) {
	if action == "session" {
		if method == fasthttp.MethodGet {
			s.handleSession(ctx)
		} else {
			methodNotAllowed(ctx)
		}
		return
	}

	if method != fasthttp.MethodPost {
		methodNotAllowed(ctx)
		return
	}

	switch action {
	case "request-code":
		s.handleRequestCode(ctx)
	case "verify-code":
		s.handleVerifyCode(ctx)
	case "verify-password":
		s.handleVerifyPassword(ctx)
	case "logout":
		s.handleLogout(ctx)
	default:
		notFound(ctx)
	}
}

func notFound(ctx *fasthttp.RequestCtx) {
	writeErrorCode(ctx, fasthttp.StatusNotFound, "NOT_FOUND", "endpoint not found")
}

func methodNotAllowed(ctx *fasthttp.RequestCtx) {
	writeErrorCode(ctx, fasthttp.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
}
