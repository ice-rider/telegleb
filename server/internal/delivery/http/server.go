package http

import (
	"context"
	"fmt"
	"log/slog"

	"telegleb/internal/config"
	"telegleb/internal/core/usecase/auth"
	"telegleb/internal/core/usecase/messenger"
	"telegleb/internal/lib/jwt"

	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

type Server struct {
	server *fasthttp.Server
	cfg    *config.Config
	log    *slog.Logger
	redis  *redis.Client
	tokens *jwt.TokenManager

	requestCodeUC    *auth.RequestCodeUseCase
	verifyCodeUC     *auth.VerifyCodeUseCase
	verifyPasswordUC *auth.VerifyPasswordUseCase
	sessionUC        *auth.SessionUseCase
	logoutUC         *auth.LogoutUseCase

	loadDashboardUC *messenger.LoadDashboardUseCase
	openChatUC      *messenger.OpenChatUseCase
	sendMessageUC   *messenger.SendMessageUseCase
	streamMediaUC   *messenger.StreamMediaUseCase
}

func NewServer(
	cfg *config.Config,
	log *slog.Logger,
	rdb *redis.Client,
	tokens *jwt.TokenManager,
	requestCodeUC *auth.RequestCodeUseCase,
	verifyCodeUC *auth.VerifyCodeUseCase,
	verifyPasswordUC *auth.VerifyPasswordUseCase,
	sessionUC *auth.SessionUseCase,
	logoutUC *auth.LogoutUseCase,
	loadDashboardUC *messenger.LoadDashboardUseCase,
	openChatUC *messenger.OpenChatUseCase,
	sendMessageUC *messenger.SendMessageUseCase,
	streamMediaUC *messenger.StreamMediaUseCase,
) *Server {
	s := &Server{
		cfg:              cfg,
		log:              log,
		redis:            rdb,
		tokens:           tokens,
		requestCodeUC:    requestCodeUC,
		verifyCodeUC:     verifyCodeUC,
		verifyPasswordUC: verifyPasswordUC,
		sessionUC:        sessionUC,
		logoutUC:         logoutUC,
		loadDashboardUC:  loadDashboardUC,
		openChatUC:       openChatUC,
		sendMessageUC:    sendMessageUC,
		streamMediaUC:    streamMediaUC,
	}

	s.server = &fasthttp.Server{
		Handler:            middlewareRecover(log, middlewareLogger(log, s.setupRouter())),
		ReadBufferSize:     8192,
		WriteBufferSize:    8192,
		StreamRequestBody:  false,
		MaxRequestBodySize: 4 * 1024 * 1024,
	}

	return s
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	addr := s.Addr()
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
	return fmt.Sprintf(":%s", s.cfg.HTTP.Port)
}
