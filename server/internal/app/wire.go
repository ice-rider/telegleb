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

func provideTelegramAppID(cfg *config.Config) int {
	return cfg.Telegram.AppID
}

func provideTelegramAppHash(cfg *config.Config) string {
	return cfg.Telegram.AppHash
}

func provideTelegramProxyAddr(cfg *config.Config) string {
	return cfg.Telegram.ProxyAddr
}

func provideTelegramProxySecret(cfg *config.Config) string {
	return cfg.Telegram.ProxySecret
}

func provideJWTManager(cfg *config.Config) *jwt.TokenManager {
	return jwt.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)
}

func InitApp() (*App, error) {
	panic(wire.Build(
		config.NewConfig,
		ProvideLogger,
		ProvideRedis,

		provideTelegramAppID,
		provideTelegramAppHash,
		provideTelegramProxyAddr,
		provideTelegramProxySecret,
		provideJWTManager,

		session.NewRedisSessionRepository,

		telegram.ProviderSet,

		auth.NewRequestLoginUseCase,
		auth.NewVerifyCodeUseCase,
		auth.NewVerifyPasswordUseCase,
		auth.NewLogoutUseCase,

		messenger.NewLoadDashboardUseCase,
		messenger.NewOpenChatUseCase,
		messenger.NewSendMessageUseCase,
		messenger.NewStreamMediaChunkUseCase,

		deliveryhttp.NewServer,

		NewApp,
	))
	return nil, nil
}
