package telegram

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"time"

	"telegleb/internal/core/domain"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func (a *TelegramAdapter) InitiateLogin(ctx context.Context, session *domain.AuthSession, phoneNumber string) (string, string, int, error) {
	a.log.Info("initiating login",
		slog.String("phone", phoneNumber),
		slog.String("session_token", shortToken(session.SessionToken)),
	)

	client, err := a.GetOrCreateClient(ctx, session)
	if err != nil {
		a.log.Error("failed to get client for login", slog.String("error", err.Error()))
		return "", "", 0, fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	start := time.Now()
	res, err := client.API().AuthSendCode(ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phoneNumber,
		APIID:       a.appID,
		APIHash:     a.appHash,
		Settings:    tg.CodeSettings{AllowAppHash: true},
	})
	elapsed := time.Since(start)

	if err != nil {
		a.log.Error("failed to send auth code",
			slog.String("phone", phoneNumber),
			slog.String("error", err.Error()),
			slog.Duration("elapsed", elapsed),
		)
		return "", "", 0, wrapTelegram(err, ErrTelegramSendCode)
	}

	sent, ok := res.(*tg.AuthSentCode)
	if !ok {
		a.log.Error("unexpected response type on send code", slog.String("type", fmt.Sprintf("%T", res)))
		return "", "", 0, ErrUnexpectedCodeType
	}

	codeType := mapSentCodeType(sent.Type)
	timeout, _ := sent.GetTimeout()

	a.log.Info("auth code sent",
		slog.String("phone", phoneNumber),
		slog.Duration("elapsed", elapsed),
		slog.String("code_type", codeType),
	)
	return sent.PhoneCodeHash, codeType, timeout, nil
}

func mapSentCodeType(t tg.AuthSentCodeTypeClass) string {
	switch t.(type) {
	case *tg.AuthSentCodeTypeApp:
		return "app"
	case *tg.AuthSentCodeTypeSMS, *tg.AuthSentCodeTypeSMSWord, *tg.AuthSentCodeTypeSMSPhrase:
		return "sms"
	case *tg.AuthSentCodeTypeCall:
		return "call"
	case *tg.AuthSentCodeTypeFlashCall:
		return "flashCall"
	case *tg.AuthSentCodeTypeMissedCall:
		return "missedCall"
	case *tg.AuthSentCodeTypeFragmentSMS:
		return "fragmentSms"
	case *tg.AuthSentCodeTypeEmailCode:
		return "email"
	default:
		return "unknown"
	}
}

func (a *TelegramAdapter) SubmitCode(ctx context.Context, session *domain.AuthSession, code string) (bool, error) {
	a.log.Info("submitting auth code",
		slog.String("phone", session.PhoneNumber),
		slog.Bool("has_code_hash", session.TelegramCodeHash != ""),
	)

	client, err := a.GetOrCreateClient(ctx, session)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	res, err := client.API().AuthSignIn(ctx, &tg.AuthSignInRequest{
		PhoneNumber:   session.PhoneNumber,
		PhoneCodeHash: session.TelegramCodeHash,
		PhoneCode:     code,
	})
	if err != nil {
		if tgerr.Is(err, ErrCodeSessionPasswordNeeded) {
			a.log.Info("2FA password required", slog.String("phone", session.PhoneNumber))
			return true, nil
		}
		a.log.Error("failed to sign in with code", slog.String("error", err.Error()))
		return false, wrapTelegram(err, ErrTelegramSignIn)
	}

	if _, ok := res.(*tg.AuthAuthorization); !ok {
		return false, ErrUnexpectedSignInType
	}

	a.log.Info("sign in successful", slog.String("phone", session.PhoneNumber))
	return false, nil
}

func (a *TelegramAdapter) SubmitPassword(ctx context.Context, session *domain.AuthSession, password string) error {
	a.log.Info("submitting 2FA password", slog.String("phone", session.PhoneNumber))

	activeClient, err := a.GetClient(ctx, session.SessionToken)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrClientInitFailed, err)
	}

	authClient := auth.NewClient(activeClient.API, rand.Reader, a.appID, a.appHash)
	if _, err := authClient.Password(ctx, password); err != nil {
		a.log.Error("invalid 2FA password", slog.String("error", err.Error()))
		return wrapTelegram(err, ErrInvalid2FAPassword)
	}

	a.log.Info("2FA password verified", slog.String("phone", session.PhoneNumber))
	return nil
}

func (a *TelegramAdapter) GetMe(ctx context.Context, sessionToken string) (domain.Me, error) {
	client, err := a.GetClient(ctx, sessionToken)
	if err != nil {
		return domain.Me{}, err
	}
	return fetchMe(ctx, client.API)
}

func shortToken(token string) string {
	if len(token) <= 8 {
		return "..."
	}
	return token[:8] + "..."
}
