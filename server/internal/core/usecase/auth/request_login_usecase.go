package auth

import (
	"context"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/session"
	"telegleb/internal/lib/jwt"
	"time"
)

type RequestLoginUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
	jwtManager  *jwt.TokenManager
}

func NewRequestLoginUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository, jwtManager *jwt.TokenManager) *RequestLoginUseCase {
	return &RequestLoginUseCase{
		authRepo:    authRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
	}
}

func (uc *RequestLoginUseCase) Execute(ctx context.Context, input RequestLoginInput) (RequestLoginOutput, error) {
	if err := input.Validate(); err != nil {
		return RequestLoginOutput{}, err
	}

	token, err := uc.jwtManager.GenerateToken(input.PhoneNumber)
	if err != nil {
		return RequestLoginOutput{}, err
	}

	authSession := &domain.AuthSession{
		SessionToken: token,
		PhoneNumber:  input.PhoneNumber,
		Status:       domain.AWAITING_PHONE,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	codeHash, err := uc.authRepo.InitiateLogin(ctx, authSession, input.PhoneNumber)
	if err != nil {
		return RequestLoginOutput{}, err
	}

	authSession.TelegramCodeHash = codeHash
	authSession.Status = domain.AWAITING_CODE

	if err := uc.sessionRepo.CreateSession(ctx, authSession); err != nil {
		return RequestLoginOutput{}, err
	}

	return RequestLoginOutput{SessionToken: token}, nil
}
