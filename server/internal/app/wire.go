//go:build wireinject
// +build wireinject

package app

import (
	"telegleb/internal/adapter/repository/session"
	"telegleb/internal/adapter/telegram"
	"telegleb/internal/config"
	deliveryhttp "telegleb/internal/delivery/http"

	"telegleb/internal/core/usecase/auth"
	"telegleb/internal/core/usecase/messenger"
	"telegleb/internal/lib/jwt"

	"github.com/google/wire"
)

func provideTelegramConfig(cfg *config.Config) telegram.Config {
	return telegram.Config{
		AppID:       cfg.Telegram.AppID,
		AppHash:     cfg.Telegram.AppHash,
		ProxyAddr:   cfg.Telegram.ProxyAddr,
		ProxySecret: cfg.Telegram.ProxySecret,
	}
}

func provideJWTManager(cfg *config.Config) *jwt.TokenManager {
	return jwt.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)
}

func InitApp() (*App, error) {
	panic(wire.Build(
		config.NewConfig,
		ProvideLogger,
		ProvideRedis,

		provideTelegramConfig,
		provideJWTManager,

		session.NewRedisSessionRepository,

		telegram.ProviderSet,

		auth.NewRequestCodeUseCase,
		auth.NewVerifyCodeUseCase,
		auth.NewVerifyPasswordUseCase,
		auth.NewSessionUseCase,
		auth.NewLogoutUseCase,

		messenger.NewLoadDashboardUseCase,
		messenger.NewOpenChatUseCase,
		messenger.NewSendMessageUseCase,
		messenger.NewStreamMediaUseCase,

		deliveryhttp.NewServer,

		NewApp,
	))
	return nil, nil
}
