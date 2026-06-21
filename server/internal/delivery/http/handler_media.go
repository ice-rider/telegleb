package http

import (
	"strconv"
	"strings"

	"telegleb/internal/core/usecase/messenger"

	"github.com/valyala/fasthttp"
)

func (s *Server) handleStreamMedia(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		writeError(ctx, fasthttp.StatusBadRequest, "missing media id")
		return
	}
	mediaID := parts[len(parts)-1]

	token := string(ctx.QueryArgs().Peek("token"))
	if token == "" {
		writeError(ctx, fasthttp.StatusUnauthorized, "missing token")
		return
	}

	offsetStr := string(ctx.QueryArgs().Peek("offset"))
	limitStr := string(ctx.QueryArgs().Peek("limit"))

	offset, _ := strconv.ParseInt(offsetStr, 10, 64)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 64 * 1024
	}

	output, err := s.streamMediaUC.Execute(ctx, messenger.StreamMediaChunkInput{
		SessionToken: token,
		MediaID:      mediaID,
		Offset:       offset,
		Limit:        limit,
	})
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	ctx.SetContentType("application/octet-stream")
	ctx.SetBody(output.Chunk)
}
