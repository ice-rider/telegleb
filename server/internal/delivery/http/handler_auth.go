package http

import (
	"errors"
	"log/slog"
	"telegleb/internal/core/usecase/auth"

	"github.com/valyala/fasthttp"
)

type requestLoginRequest struct {
	PhoneNumber string
}

type requestLoginResponse struct {
	SessionToken string `json:"sessionToken"`
}

type verifyCodeRequest struct {
	SessionToken string
	Code         string
}

type verifyCodeResponse struct {
	NextStep string `json:"nextStep"`
}

type verifyPasswordRequest struct {
	SessionToken string
	Password     string
}

type verifyPasswordResponse struct {
	Status string `json:"status"`
}

type logoutRequest struct {
	SessionToken string
}

func mapAuthErr(err error) int {
	switch {
	case errors.Is(err, auth.ErrInvalidPhone),
		errors.Is(err, auth.ErrInvalidStep),
		errors.Is(err, auth.ErrInvalidSessionState):
		return fasthttp.StatusBadRequest
	case errors.Is(err, auth.ErrInvalidCredentials),
		errors.Is(err, auth.ErrSessionNotFound):
		return fasthttp.StatusUnauthorized
	default:
		return fasthttp.StatusInternalServerError
	}
}

func (s *Server) handleRequestLogin(ctx *fasthttp.RequestCtx) {
	var req requestLoginRequest
	if err := parseBody(ctx, &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	s.log.Info("request login", slog.String("phone", req.PhoneNumber))

	output, err := s.requestLoginUC.Execute(ctx, auth.RequestLoginInput{
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		s.log.Error("request login failed", slog.String("error", err.Error()))
		writeError(ctx, mapAuthErr(err), err.Error())
		return
	}

	s.log.Info("request login success")
	writeJSON(ctx, fasthttp.StatusOK, requestLoginResponse{
		SessionToken: output.SessionToken,
	})
}

func (s *Server) handleVerifyCode(ctx *fasthttp.RequestCtx) {
	var req verifyCodeRequest
	if err := parseBody(ctx, &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	s.log.Info("verify code")

	output, err := s.verifyCodeUC.Execute(ctx, auth.VerifyCodeInput{
		SessionToken: req.SessionToken,
		Code:         req.Code,
	})
	if err != nil {
		s.log.Error("verify code failed", slog.String("error", err.Error()))
		writeError(ctx, mapAuthErr(err), err.Error())
		return
	}

	s.log.Info("verify code success", slog.String("nextStep", string(output.NextStep)))
	writeJSON(ctx, fasthttp.StatusOK, verifyCodeResponse{
		NextStep: string(output.NextStep),
	})
}

func (s *Server) handleVerifyPassword(ctx *fasthttp.RequestCtx) {
	var req verifyPasswordRequest
	if err := parseBody(ctx, &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	s.log.Info("verify password")

	output, err := s.verifyPasswordUC.Execute(ctx, auth.VerifyPasswordInput{
		SessionToken: req.SessionToken,
		Password:     req.Password,
	})
	if err != nil {
		s.log.Error("verify password failed", slog.String("error", err.Error()))
		writeError(ctx, mapAuthErr(err), err.Error())
		return
	}

	s.log.Info("verify password success")
	writeJSON(ctx, fasthttp.StatusOK, verifyPasswordResponse{
		Status: output.Status,
	})
}

func (s *Server) handleLogout(ctx *fasthttp.RequestCtx) {
	var req logoutRequest
	if err := parseBody(ctx, &req); err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, "invalid request body")
		return
	}

	s.log.Info("logout")

	if err := s.logoutUC.Execute(ctx, auth.LogoutInput{
		SessionToken: req.SessionToken,
	}); err != nil {
		s.log.Error("logout failed", slog.String("error", err.Error()))
		writeError(ctx, mapAuthErr(err), err.Error())
		return
	}

	s.log.Info("logout success")
	writeJSON(ctx, fasthttp.StatusOK, map[string]string{})
}
