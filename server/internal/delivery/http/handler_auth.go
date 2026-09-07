package http

import (
	"telegleb/internal/core/usecase/auth"

	"github.com/valyala/fasthttp"
)

type requestCodeRequest struct {
	Phone string `json:"phone"`
}

type requestCodeResponse struct {
	SessionToken string `json:"sessionToken"`
	NextStep     string `json:"nextStep"`
	CodeType     string `json:"codeType"`
	Timeout      int    `json:"timeout,omitempty"`
}

type verifyCodeRequest struct {
	Code string `json:"code"`
}

type verifyPasswordRequest struct {
	Password string `json:"password"`
}

type authStepResponse struct {
	NextStep string `json:"nextStep"`
	Me       *meDTO `json:"me,omitempty"`
}

type sessionResponse struct {
	Status string `json:"status"`
	Me     meDTO  `json:"me"`
}

func (s *Server) handleRequestCode(ctx *fasthttp.RequestCtx) {
	var req requestCodeRequest
	if err := parseBody(ctx, &req); err != nil {
		writeErrorCode(ctx, fasthttp.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	output, err := s.requestCodeUC.Execute(ctx, auth.RequestCodeInput{Phone: req.Phone})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, requestCodeResponse{
		SessionToken: output.SessionToken,
		NextStep:     string(output.NextStep),
		CodeType:     output.CodeType,
		Timeout:      output.Timeout,
	})
}

func (s *Server) handleVerifyCode(ctx *fasthttp.RequestCtx) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	var req verifyCodeRequest
	if err := parseBody(ctx, &req); err != nil {
		writeErrorCode(ctx, fasthttp.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	output, err := s.verifyCodeUC.Execute(ctx, auth.VerifyCodeInput{
		SessionToken: token,
		Code:         req.Code,
	})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	resp := authStepResponse{NextStep: string(output.NextStep)}
	if output.Me != nil {
		me := mapMe(*output.Me)
		resp.Me = &me
	}
	writeJSON(ctx, fasthttp.StatusOK, resp)
}

func (s *Server) handleVerifyPassword(ctx *fasthttp.RequestCtx) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	var req verifyPasswordRequest
	if err := parseBody(ctx, &req); err != nil {
		writeErrorCode(ctx, fasthttp.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	output, err := s.verifyPasswordUC.Execute(ctx, auth.VerifyPasswordInput{
		SessionToken: token,
		Password:     req.Password,
	})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	me := mapMe(output.Me)
	writeJSON(ctx, fasthttp.StatusOK, authStepResponse{
		NextStep: string(output.NextStep),
		Me:       &me,
	})
}

func (s *Server) handleSession(ctx *fasthttp.RequestCtx) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	output, err := s.sessionUC.Execute(ctx, auth.SessionInput{SessionToken: token})
	if err != nil {
		s.writeError(ctx, err)
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, sessionResponse{
		Status: "authorized",
		Me:     mapMe(output.Me),
	})
}

func (s *Server) handleLogout(ctx *fasthttp.RequestCtx) {
	token, ok := s.bearerToken(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	if err := s.logoutUC.Execute(ctx, auth.LogoutInput{SessionToken: token}); err != nil {
		s.writeError(ctx, err)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

func unauthorized(ctx *fasthttp.RequestCtx) {
	writeErrorCode(ctx, fasthttp.StatusUnauthorized, "SESSION_EXPIRED", "missing or invalid session token")
}
