package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/session"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
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
	sessionRepo session.SessionRepository
	log         *slog.Logger

	mu         sync.RWMutex
	clientPool map[string]*ActiveClient
}

func NewTelegramAdapter(appID int, appHash string, sessionRepo session.SessionRepository, log *slog.Logger) *TelegramAdapter {
	return &TelegramAdapter{
		appID:       appID,
		appHash:     appHash,
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
	)

	bridge := NewSessionBridge(&authSession.TelegramSessionData, func(updateCtx context.Context, data []byte) error {
		authSession.TelegramSessionData = data
		return a.sessionRepo.UpdateSession(updateCtx, authSession)
	})

	client := telegram.NewClient(a.appID, a.appHash, telegram.Options{
		SessionStorage: bridge,
	})

	runCtx, runCancel := context.WithCancel(context.Background())
	doneChan := make(chan error, 1)
	readyChan := make(chan struct{})

	go func() {
		a.log.Info("starting telegram client.Run()")
		err := client.Run(runCtx, func(cCtx context.Context) error {
			a.log.Info("telegram client connected and ready")
			close(readyChan)
			<-cCtx.Done()
			return nil
		})
		a.log.Info("telegram client.Run() exited", slog.String("error", err.Error()))
		doneChan <- err
		close(doneChan)
	}()

	select {
	case <-readyChan:
		a.log.Info("telegram client ready")
	case err := <-doneChan:
		runCancel()
		a.log.Error("telegram client failed to start", slog.String("error", err.Error()))
		return nil, fmt.Errorf("telegram client failed to start: %w", err)
	case <-ctx.Done():
		runCancel()
		a.log.Error("context cancelled while waiting for telegram client")
		return nil, ctx.Err()
	}

	a.clientPool[authSession.SessionToken] = &ActiveClient{
		Client:    client,
		CancelRun: runCancel,
		Done:      doneChan,
	}

	return client, nil
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
