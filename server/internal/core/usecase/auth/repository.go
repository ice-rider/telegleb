package auth

import (
	"context"
	"telegleb/internal/core/domain"
)

type AuthRepository interface {
	// InitiateLogin отправляет код на телефон и возвращает telegram-phone-code-hash
	InitiateLogin(ctx context.Context, session *domain.AuthSession, phoneNumber string) (string, error)
	// SubmitCode проверяет код. Возвращает true, если нужен 2FA пароль, и false, если авторизация успешна
	SubmitCode(ctx context.Context, session *domain.AuthSession, code string) (requiresPassword bool, error error)
	// SubmitPassword проверяет 2FA пароль
	SubmitPassword(ctx context.Context, session *domain.AuthSession, password string) error
	// TerminateSession корректно закрывает соединение и очищает ресурсы
	TerminateSession(ctx context.Context, sessionToken string) error
}