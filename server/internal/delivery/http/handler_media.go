package http

import (
	"fmt"
	"strconv"
	"strings"

	"telegleb/internal/core/usecase/messenger"

	"github.com/valyala/fasthttp"
)

const (
	// Telegram требует, чтобы offset был кратен 4096, limit был кратен 4096 и
	// делил мегабайт нацело. 512 КиБ удовлетворяет всем трём условиям.
	mediaAlignment = 4096
	mediaChunkSize = 512 * 1024
	// Потолок на ответ без Range: столько байт мы готовы собрать в память.
	mediaFullLimit = 16 * 1024 * 1024
)

func (s *Server) handleStreamMedia(ctx *fasthttp.RequestCtx, mediaRef string) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	rangeStart, hasRange := parseRangeStart(string(ctx.Request.Header.Peek("Range")))

	if !hasRange {
		s.serveWholeMedia(ctx, token, mediaRef)
		return
	}

	// Смещение выравнивается вниз до границы блока, лишний префикс срезается
	// уже из полученного чанка.
	aligned := rangeStart - rangeStart%mediaAlignment
	chunk, err := s.streamMediaUC.Execute(ctx, messenger.StreamMediaInput{
		SessionToken: token,
		MediaRef:     mediaRef,
		Offset:       aligned,
		Limit:        mediaChunkSize,
	})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	body := chunk.Bytes
	if skip := int(rangeStart - aligned); skip > 0 && skip < len(body) {
		body = body[skip:]
	} else if skip >= len(body) {
		body = nil
	}

	ctx.SetStatusCode(fasthttp.StatusPartialContent)
	ctx.SetContentType(contentType(chunk.MimeType))
	ctx.Response.Header.Set("Accept-Ranges", "bytes")
	if chunk.Total > 0 {
		end := rangeStart + int64(len(body)) - 1
		ctx.Response.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rangeStart, end, chunk.Total))
	}
	ctx.SetBody(body)
}

func (s *Server) serveWholeMedia(ctx *fasthttp.RequestCtx, token, mediaRef string) {
	var (
		body     []byte
		offset   int64
		mimeType string
		total    int64
	)

	for {
		chunk, err := s.streamMediaUC.Execute(ctx, messenger.StreamMediaInput{
			SessionToken: token,
			MediaRef:     mediaRef,
			Offset:       offset,
			Limit:        mediaChunkSize,
		})
		if err != nil {
			s.writeError(ctx, err)
			return
		}

		mimeType, total = chunk.MimeType, chunk.Total
		body = append(body, chunk.Bytes...)
		offset += int64(len(chunk.Bytes))

		done := len(chunk.Bytes) < mediaChunkSize ||
			(total > 0 && offset >= total) ||
			len(body) >= mediaFullLimit
		if done {
			break
		}
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType(contentType(mimeType))
	ctx.Response.Header.Set("Accept-Ranges", "bytes")
	ctx.Response.Header.Set("Cache-Control", "private, max-age=86400")
	ctx.SetBody(body)
}

// parseRangeStart разбирает только начало диапазона: конец мы всё равно
// определяем размером чанка, который отдаёт Telegram.
func parseRangeStart(header string) (int64, bool) {
	if !strings.HasPrefix(header, "bytes=") {
		return 0, false
	}
	spec := strings.TrimPrefix(header, "bytes=")
	if idx := strings.IndexByte(spec, ','); idx >= 0 {
		spec = spec[:idx]
	}
	startStr, _, found := strings.Cut(spec, "-")
	if !found {
		return 0, false
	}
	start, err := strconv.ParseInt(strings.TrimSpace(startStr), 10, 64)
	if err != nil || start < 0 {
		return 0, false
	}
	return start, true
}

func contentType(mimeType string) string {
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}
