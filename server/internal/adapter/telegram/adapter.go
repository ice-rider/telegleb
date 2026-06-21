package telegram

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/session"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ActiveClient struct {
	Client    *telegram.Client
	API       *tg.Client
	CancelRun context.CancelFunc
	Done      <-chan error
}

type TelegramAdapter struct {
	appID       int
	appHash     string
	proxyAddr   string
	proxySecret string
	sessionRepo session.SessionRepository
	log         *slog.Logger

	mu         sync.RWMutex
	clientPool map[string]*ActiveClient
}

func NewTelegramAdapter(appID int, appHash string, proxyAddr string, proxySecret string, sessionRepo session.SessionRepository, log *slog.Logger) *TelegramAdapter {
	return &TelegramAdapter{
		appID:       appID,
		appHash:     appHash,
		proxyAddr:   proxyAddr,
		proxySecret: proxySecret,
		sessionRepo: sessionRepo,
		log:         log,
		clientPool:  make(map[string]*ActiveClient),
	}
}

func (a *TelegramAdapter) GetClient(ctx context.Context, sessionToken string) (*ActiveClient, error) {
	a.mu.RLock()
	client, exists := a.clientPool[sessionToken]
	a.mu.RUnlock()

	if exists {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-client.Done:
			if err != nil {
				a.log.Warn("client done with error, reinitializing", slog.String("error", err.Error()))
				a.mu.Lock()
				delete(a.clientPool, sessionToken)
				a.mu.Unlock()
				return a.initAndFetchClient(ctx, sessionToken)
			}
		default:
			return client, nil
		}
	}

	return a.initAndFetchClient(ctx, sessionToken)
}

func (a *TelegramAdapter) initAndFetchClient(ctx context.Context, sessionToken string) (*ActiveClient, error) {
	a.log.Info("init and fetch client", slog.String("token", sessionToken[:8]+"..."))

	authSession, err := a.sessionRepo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		a.log.Error("session not found", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", ErrSessionNotFound, err)
	}

	_, err = a.GetOrCreateClient(ctx, authSession)
	if err != nil {
		return nil, err
	}

	a.mu.RLock()
	active := a.clientPool[sessionToken]
	a.mu.RUnlock()
	return active, nil
}

func (a *TelegramAdapter) buildClientOptions(authSession *domain.AuthSession) telegram.Options {
	bridge := NewSessionBridge(&authSession.TelegramSessionData, func(updateCtx context.Context, data []byte) error {
		authSession.TelegramSessionData = data
		return a.sessionRepo.UpdateSession(updateCtx, authSession)
	})

	opts := telegram.Options{
		SessionStorage: bridge,
		OnDead: func(err error) {
			a.log.Warn("telegram connection dead", slog.String("error", err.Error()))
		},
		Logger: a.buildZapLogger(),
	}

	if a.proxyAddr != "" && a.proxySecret != "" {
		a.log.Info("configuring MTProxy resolver", slog.String("addr", a.proxyAddr))
		secretBytes, err := hex.DecodeString(a.proxySecret)
		if err != nil {
			a.log.Error("failed to decode MTProxy secret hex", slog.String("error", err.Error()))
		} else {
			resolver, err := dcs.MTProxy(a.proxyAddr, secretBytes, dcs.MTProxyOptions{})
			if err != nil {
				a.log.Error("failed to create MTProxy resolver", slog.String("error", err.Error()))
			} else {
				opts.Resolver = resolver
			}
		}
	}

	return opts
}

func (a *TelegramAdapter) GetOrCreateClient(ctx context.Context, authSession *domain.AuthSession) (*telegram.Client, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if active, exists := a.clientPool[authSession.SessionToken]; exists {
		a.log.Info("using existing client", slog.String("token", authSession.SessionToken[:8]+"..."))
		return active.Client, nil
	}

	a.log.Info("creating new telegram client",
		slog.String("token", authSession.SessionToken[:8]+"..."),
		slog.Int("appID", a.appID),
		slog.Bool("hasProxy", a.proxyAddr != ""),
	)

	opts := a.buildClientOptions(authSession)
	client := telegram.NewClient(a.appID, a.appHash, opts)

	runCtx, runCancel := context.WithCancel(context.Background())
	doneChan := make(chan error, 1)
	readyChan := make(chan *tg.Client, 1)

	go func() {
		a.log.Info("starting telegram client.Run()")
		err := client.Run(runCtx, func(cCtx context.Context) error {
			a.log.Info("telegram client.Run callback fired")
			readyChan <- client.API()
			a.log.Info("telegram client callback waiting for shutdown")
			<-cCtx.Done()
			return nil
		})
		a.log.Info("telegram client.Run() exited", slog.String("error", err.Error()))
		doneChan <- err
		close(doneChan)
	}()

	select {
	case api := <-readyChan:
		a.log.Info("telegram client ready, storing in pool")
		a.clientPool[authSession.SessionToken] = &ActiveClient{
			Client:    client,
			API:       api,
			CancelRun: runCancel,
			Done:      doneChan,
		}
		return client, nil
	case err := <-doneChan:
		runCancel()
		a.log.Error("telegram client failed to start", slog.String("error", err.Error()))
		return nil, fmt.Errorf("telegram client failed to start: %w", err)
	case <-ctx.Done():
		runCancel()
		a.log.Error("context cancelled while waiting for telegram client")
		return nil, ctx.Err()
	}
}

func (a *TelegramAdapter) buildZapLogger() *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zapcore.DebugLevel,
	)
	return zap.New(core)
}

func (a *TelegramAdapter) TerminateSession(ctx context.Context, sessionToken string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	active, exists := a.clientPool[sessionToken]
	if !exists {
		return nil
	}

	a.log.Info("terminating session", slog.String("token", sessionToken[:8]+"..."))
	active.CancelRun()
	delete(a.clientPool, sessionToken)

	return nil
}
