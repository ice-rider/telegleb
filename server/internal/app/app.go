package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"telegleb/internal/config"
	deliveryhttp "telegleb/internal/delivery/http"
	"telegleb/internal/lib/logger"

	"github.com/redis/go-redis/v9"
)

type App struct {
	Cfg *config.Config
	Log *slog.Logger
	Rdb *redis.Client

	httpServer *deliveryhttp.Server
}

func NewApp(
	cfg *config.Config,
	log *slog.Logger,
	rdb *redis.Client,
	httpServer *deliveryhttp.Server,
) *App {
	return &App{
		Cfg:       cfg,
		Log:       log,
		Rdb:       rdb,
		httpServer: httpServer,
	}
}

func ProvideLogger(cfg *config.Config) *slog.Logger {
	return logger.SetupLogger(cfg.Log.Level, cfg.Log.JSON)
}

func ProvideRedis(cfg *config.Config, log *slog.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Redis.Timeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	log.Info("connected to redis successfully", slog.String("host", cfg.Redis.Host))
	return rdb, nil
}

func (a *App) Run() error {
	a.Log.Info("application started successfully", slog.String("env", a.Cfg.Env))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.httpServer.ListenAndServe(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		a.Log.Error("http server error", slog.String("error", err.Error()))
	case sign := <-stop:
		a.Log.Info("stopping application gracefully", slog.String("signal", sign.String()))
	}

	cancel()
	a.httpServer.Shutdown()

	if err := a.Rdb.Close(); err != nil {
		a.Log.Error("failed to close redis client", slog.String("error", err.Error()))
	}

	a.Log.Info("application stopped completely")
	return nil
}
