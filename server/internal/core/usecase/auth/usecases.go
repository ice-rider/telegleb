package auth

import (
	"context"
	"errors"
	"time"

	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/session"
	"telegleb/internal/lib/jwt"
)

const sessionTTL = 24 * time.Hour

type RequestCodeUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
	jwtManager  *jwt.TokenManager
}

func NewRequestCodeUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository, jwtManager *jwt.TokenManager) *RequestCodeUseCase {
	return &RequestCodeUseCase{authRepo: authRepo, sessionRepo: sessionRepo, jwtManager: jwtManager}
}

func (uc *RequestCodeUseCase) Execute(ctx context.Context, input RequestCodeInput) (RequestCodeOutput, error) {
	if err := input.Validate(); err != nil {
		return RequestCodeOutput{}, err
	}

	token, err := uc.jwtManager.GenerateToken(input.Phone)
	if err != nil {
		return RequestCodeOutput{}, err
	}

	authSession := &domain.AuthSession{
		SessionToken: token,
		PhoneNumber:  input.Phone,
		Status:       domain.StatusAwaitingCode,
		ExpiresAt:    time.Now().Add(sessionTTL),
	}

	codeHash, codeType, timeout, err := uc.authRepo.InitiateLogin(ctx, authSession, input.Phone)
	if err != nil {
		return RequestCodeOutput{}, err
	}

	authSession.TelegramCodeHash = codeHash
	if err := uc.sessionRepo.CreateSession(ctx, authSession); err != nil {
		return RequestCodeOutput{}, err
	}

	return RequestCodeOutput{
		SessionToken: token,
		NextStep:     domain.NextStepCode,
		CodeType:     codeType,
		Timeout:      timeout,
	}, nil
}

type VerifyCodeUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
}

func NewVerifyCodeUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository) *VerifyCodeUseCase {
	return &VerifyCodeUseCase{authRepo: authRepo, sessionRepo: sessionRepo}
}

func (uc *VerifyCodeUseCase) Execute(ctx context.Context, input VerifyCodeInput) (VerifyCodeOutput, error) {
	if err := input.Validate(); err != nil {
		return VerifyCodeOutput{}, err
	}

	authSession, err := uc.sessionRepo.GetSessionByToken(ctx, input.SessionToken)
	if err != nil {
		return VerifyCodeOutput{}, errors.Join(ErrSessionNotFound, err)
	}
	if authSession.Status != domain.StatusAwaitingCode {
		return VerifyCodeOutput{}, ErrInvalidSessionState
	}

	requiresPassword, err := uc.authRepo.SubmitCode(ctx, authSession, input.Code)
	if err != nil {
		return VerifyCodeOutput{}, err
	}

	if requiresPassword {
		authSession.Status = domain.StatusAwaitingPassword
		if err := uc.sessionRepo.UpdateSession(ctx, authSession); err != nil {
			return VerifyCodeOutput{}, err
		}
		return VerifyCodeOutput{NextStep: domain.NextStepPassword}, nil
	}

	authSession.Status = domain.StatusAuthorized
	authSession.ExpiresAt = time.Now().Add(sessionTTL)
	if err := uc.sessionRepo.UpdateSession(ctx, authSession); err != nil {
		return VerifyCodeOutput{}, err
	}

	me, err := uc.authRepo.GetMe(ctx, input.SessionToken)
	if err != nil {
		return VerifyCodeOutput{}, err
	}
	return VerifyCodeOutput{NextStep: domain.NextStepDone, Me: &me}, nil
}

type VerifyPasswordUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
}

func NewVerifyPasswordUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository) *VerifyPasswordUseCase {
	return &VerifyPasswordUseCase{authRepo: authRepo, sessionRepo: sessionRepo}
}

func (uc *VerifyPasswordUseCase) Execute(ctx context.Context, input VerifyPasswordInput) (VerifyPasswordOutput, error) {
	if err := input.Validate(); err != nil {
		return VerifyPasswordOutput{}, err
	}

	authSession, err := uc.sessionRepo.GetSessionByToken(ctx, input.SessionToken)
	if err != nil {
		return VerifyPasswordOutput{}, errors.Join(ErrSessionNotFound, err)
	}
	if authSession.Status != domain.StatusAwaitingPassword {
		return VerifyPasswordOutput{}, ErrInvalidSessionState
	}

	if err := uc.authRepo.SubmitPassword(ctx, authSession, input.Password); err != nil {
		return VerifyPasswordOutput{}, err
	}

	authSession.Status = domain.StatusAuthorized
	authSession.ExpiresAt = time.Now().Add(sessionTTL)
	if err := uc.sessionRepo.UpdateSession(ctx, authSession); err != nil {
		return VerifyPasswordOutput{}, err
	}

	me, err := uc.authRepo.GetMe(ctx, input.SessionToken)
	if err != nil {
		return VerifyPasswordOutput{}, err
	}
	return VerifyPasswordOutput{NextStep: domain.NextStepDone, Me: me}, nil
}

// SessionUseCase проверяет живость токена при старте клиента. Без этой
// проверки клиент, восстановивший протухший токен из localStorage, попадает в
// интерфейс и получает ошибку на первом же запросе.
type SessionUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
}

func NewSessionUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository) *SessionUseCase {
	return &SessionUseCase{authRepo: authRepo, sessionRepo: sessionRepo}
}

func (uc *SessionUseCase) Execute(ctx context.Context, input SessionInput) (SessionOutput, error) {
	authSession, err := uc.sessionRepo.GetSessionByToken(ctx, input.SessionToken)
	if err != nil {
		return SessionOutput{}, errors.Join(ErrSessionNotFound, err)
	}
	if authSession.Status != domain.StatusAuthorized {
		return SessionOutput{}, ErrSessionNotFound
	}

	me, err := uc.authRepo.GetMe(ctx, input.SessionToken)
	if err != nil {
		return SessionOutput{}, err
	}
	return SessionOutput{Me: me}, nil
}

type LogoutUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
}

func NewLogoutUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{authRepo: authRepo, sessionRepo: sessionRepo}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	if err := uc.authRepo.TerminateSession(ctx, input.SessionToken); err != nil {
		return err
	}
	return uc.sessionRepo.DeleteSession(ctx, input.SessionToken)
}
