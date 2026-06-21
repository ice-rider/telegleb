package telegram

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"telegleb/internal/core/domain"
	"time"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func (a *TelegramAdapter) InitiateLogin(ctx context.Context, session *domain.AuthSession, phoneNumber string) (string, error) {
	a.log.Info("initiating login", slog.String("phone", phoneNumber))

	client, err := a.GetOrCreateClient(ctx, session)
	if err != nil {
		a.log.Error("failed to get client for login", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	api := client.API()
	a.log.Info("sending auth code", slog.String("phone", phoneNumber))

	res, err := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phoneNumber,
		APIID:       a.appID,
		APIHash:     a.appHash,
		Settings:    tg.CodeSettings{},
	})
	if err != nil {
		a.log.Error("failed to send auth code", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %v", ErrTelegramSendCode, err)
	}

	switch codeType := res.(type) {
	case *tg.AuthSentCode:
		a.log.Info("auth code sent successfully")
		return codeType.PhoneCodeHash, nil
	default:
		a.log.Error("unexpected response type on send code")
		return "", ErrUnexpectedCodeType
	}
}

func (a *TelegramAdapter) SubmitCode(ctx context.Context, session *domain.AuthSession, code string) (bool, error) {
	a.log.Info("submitting code")

	client, err := a.GetOrCreateClient(ctx, session)
	if err != nil {
		a.log.Error("failed to get client for code submission", slog.String("error", err.Error()))
		return false, fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	api := client.API()

	res, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
		PhoneNumber:   session.PhoneNumber,
		PhoneCodeHash: session.TelegramCodeHash,
		PhoneCode:     code,
	})

	if err != nil {
		if tgerr.Is(err, ErrCodeSessionPasswordNeeded) {
			a.log.Info("2FA password required")
			return true, nil
		}
		a.log.Error("failed to sign in with code", slog.String("error", err.Error()))
		return false, fmt.Errorf("%w: %v", ErrTelegramSignIn, err)
	}

	switch res.(type) {
	case *tg.AuthAuthorization:
		a.log.Info("sign in successful")
		return false, nil
	default:
		a.log.Error("unexpected response type on sign in")
		return false, ErrUnexpectedSignInType
	}
}

func (a *TelegramAdapter) SubmitPassword(ctx context.Context, session *domain.AuthSession, password string) error {
	a.log.Info("submitting 2FA password")

	activeClient, err := a.GetClient(ctx, session.SessionToken)
	if err != nil {
		a.log.Error("failed to get client for password submission", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	authClient := auth.NewClient(activeClient.API, rand.Reader, a.appID, a.appHash)

	_, err = authClient.Password(ctx, password)
	if err != nil {
		a.log.Error("invalid 2FA password", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", ErrInvalid2FAPassword, err)
	}

	a.log.Info("2FA password verified successfully")
	session.Status = domain.AUTHORIZED
	session.ExpiresAt = time.Now().Add(24 * time.Hour)

	if err := a.sessionRepo.UpdateSession(ctx, session); err != nil {
		a.log.Error("failed to update session after password verification", slog.String("error", err.Error()))
		return fmt.Errorf("failed to update session after password verification: %w", err)
	}

	return nil
}
