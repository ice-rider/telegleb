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
	a.log.Info("initiating login",
		slog.String("phone", phoneNumber),
		slog.String("session_token", session.SessionToken[:8]+"..."),
		slog.String("status", string(session.Status)),
	)

	client, err := a.GetOrCreateClient(ctx, session)
	if err != nil {
		a.log.Error("failed to get client for login",
			slog.String("phone", phoneNumber),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	api := client.API()
	a.log.Info("sending auth code to Telegram",
		slog.String("phone", phoneNumber),
		slog.Int("app_id", a.appID),
		slog.Bool("has_proxy", a.proxyAddr != ""),
	)

	start := time.Now()
	res, err := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phoneNumber,
		APIID:       a.appID,
		APIHash:     a.appHash,
		Settings: tg.CodeSettings{
			AllowAppHash: true,
		},
	})
	elapsed := time.Since(start)

	if err != nil {
		a.log.Error("failed to send auth code",
			slog.String("phone", phoneNumber),
			slog.String("error", err.Error()),
			slog.Duration("elapsed", elapsed),
		)
		return "", fmt.Errorf("%w: %v", ErrTelegramSendCode, err)
	}

	switch codeType := res.(type) {
	case *tg.AuthSentCode:
		deliveryMethod := fmt.Sprintf("%T", codeType.Type)
		a.log.Info("auth code sent successfully",
			slog.String("phone", phoneNumber),
			slog.Duration("elapsed", elapsed),
			slog.String("delivery_method", deliveryMethod),
			slog.String("phone_code_hash", codeType.PhoneCodeHash),
		)
		return codeType.PhoneCodeHash, nil
	default:
		a.log.Error("unexpected response type on send code",
			slog.String("phone", phoneNumber),
			slog.String("response_type", fmt.Sprintf("%T", res)),
		)
		return "", ErrUnexpectedCodeType
	}
}

func (a *TelegramAdapter) SubmitCode(ctx context.Context, session *domain.AuthSession, code string) (bool, error) {
	a.log.Info("submitting auth code",
		slog.String("phone", session.PhoneNumber),
		slog.Int("code_length", len(code)),
		slog.Bool("has_code_hash", session.TelegramCodeHash != ""),
	)

	client, err := a.GetOrCreateClient(ctx, session)
	if err != nil {
		a.log.Error("failed to get client for code submission",
			slog.String("phone", session.PhoneNumber),
			slog.String("error", err.Error()),
		)
		return false, fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	api := client.API()

	start := time.Now()
	res, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
		PhoneNumber:   session.PhoneNumber,
		PhoneCodeHash: session.TelegramCodeHash,
		PhoneCode:     code,
	})
	elapsed := time.Since(start)

	if err != nil {
		if tgerr.Is(err, ErrCodeSessionPasswordNeeded) {
			a.log.Info("2FA password required",
				slog.String("phone", session.PhoneNumber),
				slog.Duration("elapsed", elapsed),
			)
			return true, nil
		}
		a.log.Error("failed to sign in with code",
			slog.String("phone", session.PhoneNumber),
			slog.String("error", err.Error()),
			slog.Duration("elapsed", elapsed),
		)
		return false, fmt.Errorf("%w: %v", ErrTelegramSignIn, err)
	}

	switch r := res.(type) {
	case *tg.AuthAuthorization:
		a.log.Info("sign in successful",
			slog.String("phone", session.PhoneNumber),
			slog.Duration("elapsed", elapsed),
			slog.Any("user_id", r.User),
		)
		return false, nil
	default:
		a.log.Error("unexpected response type on sign in",
			slog.String("phone", session.PhoneNumber),
			slog.String("response_type", fmt.Sprintf("%T", res)),
			slog.Duration("elapsed", elapsed),
		)
		return false, ErrUnexpectedSignInType
	}
}

func (a *TelegramAdapter) SubmitPassword(ctx context.Context, session *domain.AuthSession, password string) error {
	a.log.Info("submitting 2FA password",
		slog.String("phone", session.PhoneNumber),
		slog.Int("password_length", len(password)),
	)

	activeClient, err := a.GetClient(ctx, session.SessionToken)
	if err != nil {
		a.log.Error("failed to get client for password submission",
			slog.String("phone", session.PhoneNumber),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	authClient := auth.NewClient(activeClient.API, rand.Reader, a.appID, a.appHash)

	start := time.Now()
	_, err = authClient.Password(ctx, password)
	elapsed := time.Since(start)

	if err != nil {
		a.log.Error("invalid 2FA password",
			slog.String("phone", session.PhoneNumber),
			slog.String("error", err.Error()),
			slog.Duration("elapsed", elapsed),
		)
		return fmt.Errorf("%w: %v", ErrInvalid2FAPassword, err)
	}

	a.log.Info("2FA password verified successfully",
		slog.String("phone", session.PhoneNumber),
		slog.Duration("elapsed", elapsed),
	)
	session.Status = domain.AUTHORIZED
	session.ExpiresAt = time.Now().Add(24 * time.Hour)

	if err := a.sessionRepo.UpdateSession(ctx, session); err != nil {
		a.log.Error("failed to update session after password verification",
			slog.String("phone", session.PhoneNumber),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to update session after password verification: %w", err)
	}

	a.log.Info("session updated after 2FA verification",
		slog.String("phone", session.PhoneNumber),
		slog.Time("expires_at", session.ExpiresAt),
	)

	return nil
}
