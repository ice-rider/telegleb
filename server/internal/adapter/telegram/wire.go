package telegram

import (
	"github.com/google/wire"
	"telegleb/internal/core/usecase/auth"
)

var ProviderSet = wire.NewSet(
	NewTelegramAdapter,
	NewTelegramChatRepository,
	NewTelegramMessageRepository,
	NewTelegramMediaRepository,

	// Связываем конкретную реализацию (*TelegramAdapter) с интерфейсом (auth.AuthRepository)
	wire.Bind(new(auth.AuthRepository), new(*TelegramAdapter)),
)
