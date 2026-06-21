package auth

import (
	"context"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/session"
)

type VerifyCodeUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
}

func NewVerifyCodeUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository) *VerifyCodeUseCase {
	return &VerifyCodeUseCase{
		authRepo:    authRepo,
		sessionRepo: sessionRepo,
	}
}

func (uc *VerifyCodeUseCase) Execute(ctx context.Context, input VerifyCodeInput) (VerifyCodeOutput, error) {
	if err := input.Validate(); err != nil {
		return VerifyCodeOutput{}, err
	}

	authSession, err := uc.sessionRepo.GetSessionByToken(ctx, input.SessionToken)
	if err != nil {
		return VerifyCodeOutput{}, err
	}

	if authSession.Status != domain.AWAITING_CODE {
		return VerifyCodeOutput{}, ErrInvalidSessionState
	}

	requiresPassword, err := uc.authRepo.SubmitCode(ctx, authSession, input.Code)
	if err != nil {
		return VerifyCodeOutput{}, err
	}

	var nextStep domain.NextStep
	if requiresPassword {
		authSession.Status = domain.AWAITING_PASSWORD
		nextStep = domain.NextStepAwaitingPassword
	} else {
		authSession.Status = domain.AUTHORIZED
		nextStep = domain.NextStepAuthorized
	}

	if err := uc.sessionRepo.UpdateSession(ctx, authSession); err != nil {
		return VerifyCodeOutput{}, err
	}

	return VerifyCodeOutput{NextStep: nextStep}, nil
}
