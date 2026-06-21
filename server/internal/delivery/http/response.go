package http

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func writeJSON(ctx *fasthttp.RequestCtx, statusCode int, data interface{}) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(data); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error":"failed to encode response"}`)
	}
}

func writeError(ctx *fasthttp.RequestCtx, statusCode int, message string) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	ctx.SetBodyString(`{"error":"` + message + `"}`)
}

func parseBody(ctx *fasthttp.RequestCtx, dst interface{}) error {
	return json.Unmarshal(ctx.PostBody(), dst)
}
