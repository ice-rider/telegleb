package telegram

import (
	"telegleb/internal/core/usecase/auth"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewTelegramAdapter,
	NewTelegramChatRepository,
	NewTelegramMessageRepository,
	NewTelegramMediaRepository,

	// Связываем конкретную реализацию (*TelegramAdapter) с интерфейсом (auth.AuthRepository)
	wire.Bind(new(auth.AuthRepository), new(*TelegramAdapter)),
)
