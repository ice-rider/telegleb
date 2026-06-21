package auth

import (
	"strings"
	"telegleb/internal/core/domain"
)

type RequestLoginInput struct {
	PhoneNumber string
}

func (i RequestLoginInput) Validate() error {
	phone := strings.TrimSpace(i.PhoneNumber)
	if phone == "" || !strings.HasPrefix(phone, "+") {
		return ErrInvalidPhone
	}
	return nil
}

type RequestLoginOutput struct {
	SessionToken string
}

type VerifyCodeInput struct {
	SessionToken string
	Code         string
}

func (i VerifyCodeInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionState
	}
	if strings.TrimSpace(i.Code) == "" {
		return ErrInvalidStep
	}
	return nil
}

type VerifyCodeOutput struct {
	NextStep domain.NextStep
}

type VerifyPasswordInput struct {
	SessionToken string
	Password     string
}

func (i VerifyPasswordInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionState
	}
	if strings.TrimSpace(i.Password) == "" {
		return ErrInvalidStep
	}
	return nil
}

type VerifyPasswordOutput struct {
	Status string
}

type LogoutInput struct {
	SessionToken string
}

func (i LogoutInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" {
		return ErrInvalidSessionState
	}
	return nil
}