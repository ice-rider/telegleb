package auth

import (
	"context"
	"telegleb/internal/core/usecase/session"
)

type LogoutUseCase struct {
	authRepo    AuthRepository
	sessionRepo session.SessionRepository
}

func NewLogoutUseCase(authRepo AuthRepository, sessionRepo session.SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{
		authRepo:    authRepo,
		sessionRepo: sessionRepo,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	_, err := uc.sessionRepo.GetSessionByToken(ctx, input.SessionToken)
	if err != nil {
		return ErrSessionNotFound
	}

	if err := uc.authRepo.TerminateSession(ctx, input.SessionToken); err != nil {
		return err
	}

	if err := uc.sessionRepo.DeleteSession(ctx, input.SessionToken); err != nil {
		return err
	}

	return nil
}
