package auth

import (
	"context"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/session"
)

type VerifyPasswordUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
}

func NewVerifyPasswordUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository) *VerifyPasswordUseCase {
	return &VerifyPasswordUseCase{
		authRepo:    authRepo,
		sessionRepo: sessionRepo,
	}
}

func (uc *VerifyPasswordUseCase) Execute(ctx context.Context, input VerifyPasswordInput) (VerifyPasswordOutput, error) {
	if err := input.Validate(); err != nil {
		return VerifyPasswordOutput{}, err
	}
	
	authSession, err := uc.sessionRepo.GetSessionByToken(ctx, input.SessionToken)
	if err != nil {
		return VerifyPasswordOutput{}, err
	}

	if authSession.Status != domain.AWAITING_PASSWORD {
		return VerifyPasswordOutput{}, ErrInvalidSessionState
	}

	err = uc.authRepo.SubmitPassword(ctx, authSession, input.Password)
	if err != nil {
		return VerifyPasswordOutput{}, err
	}

	authSession.Status = domain.AUTHORIZED
	if err := uc.sessionRepo.UpdateSession(ctx, authSession); err != nil {
		return VerifyPasswordOutput{}, err
	}

	return VerifyPasswordOutput{Status: "OK"}, nil
}
