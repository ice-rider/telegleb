package http

import (
	"context"
	"fmt"
	"log/slog"

	"telegleb/internal/config"
	"telegleb/internal/core/usecase/auth"
	"telegleb/internal/core/usecase/messenger"

	"github.com/valyala/fasthttp"
)

type Server struct {
	server *fasthttp.Server
	cfg    *config.Config
	log    *slog.Logger

	requestLoginUC   *auth.RequestLoginUseCase
	verifyCodeUC     *auth.VerifyCodeUseCase
	verifyPasswordUC *auth.VerifyPasswordUseCase
	logoutUC         *auth.LogoutUseCase

	loadDashboardUC *messenger.LoadDashboardUseCase
	openChatUC      *messenger.OpenChatUseCase
	sendMessageUC   *messenger.SendMessageUseCase
	streamMediaUC   *messenger.StreamMediaChunkUseCase
}

func NewServer(
	cfg *config.Config,
	log *slog.Logger,
	requestLoginUC *auth.RequestLoginUseCase,
	verifyCodeUC *auth.VerifyCodeUseCase,
	verifyPasswordUC *auth.VerifyPasswordUseCase,
	logoutUC *auth.LogoutUseCase,
	loadDashboardUC *messenger.LoadDashboardUseCase,
	openChatUC *messenger.OpenChatUseCase,
	sendMessageUC *messenger.SendMessageUseCase,
	streamMediaUC *messenger.StreamMediaChunkUseCase,
) *Server {
	s := &Server{
		cfg:              cfg,
		log:              log,
		requestLoginUC:   requestLoginUC,
		verifyCodeUC:     verifyCodeUC,
		verifyPasswordUC: verifyPasswordUC,
		logoutUC:         logoutUC,
		loadDashboardUC:  loadDashboardUC,
		openChatUC:       openChatUC,
		sendMessageUC:    sendMessageUC,
		streamMediaUC:    streamMediaUC,
	}

	handler := s.setupRouter()
	s.server = &fasthttp.Server{
		Handler:            middlewareRecover(middlewareCORS(handler)),
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		DisableHeaderNamesNormalizing: true,
	}

	return s
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	addr := ":8080"
	s.log.Info("http server starting", slog.String("addr", addr))

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe(addr)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.Shutdown()
		return ctx.Err()
	}
}

func (s *Server) Shutdown() {
	s.log.Info("http server shutting down")
	if err := s.server.Shutdown(); err != nil {
		s.log.Error("http server shutdown error", slog.String("error", err.Error()))
	}
}

func (s *Server) Addr() string {
	return fmt.Sprintf(":%s", "8080")
}
