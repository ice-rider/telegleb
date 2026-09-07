package auth

import (
	"context"

	"telegleb/internal/core/domain"
)

type AuthRepository interface {
	// InitiateLogin отправляет код на телефон и возвращает telegram-phone-code-hash
	// вместе со способом доставки кода.
	InitiateLogin(ctx context.Context, session *domain.AuthSession, phoneNumber string) (hash string, codeType string, timeout int, err error)
	// SubmitCode проверяет код. Возвращает true, если нужен 2FA пароль.
	SubmitCode(ctx context.Context, session *domain.AuthSession, code string) (requiresPassword bool, err error)
	// SubmitPassword проверяет 2FA пароль.
	SubmitPassword(ctx context.Context, session *domain.AuthSession, password string) error
	// GetMe возвращает профиль авторизованного пользователя.
	GetMe(ctx context.Context, sessionToken string) (domain.Me, error)
	// TerminateSession закрывает соединение и очищает ресурсы.
	TerminateSession(ctx context.Context, sessionToken string) error
}
