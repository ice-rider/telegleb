package auth

import (
	"strings"

	"telegleb/internal/core/domain"
)

type RequestCodeInput struct {
	Phone string
}

func (i RequestCodeInput) Validate() error {
	phone := strings.TrimSpace(i.Phone)
	if !strings.HasPrefix(phone, "+") || len(phone) < 6 {
		return ErrInvalidPhone
	}
	for _, r := range phone[1:] {
		if r < '0' || r > '9' {
			return ErrInvalidPhone
		}
	}
	return nil
}

type RequestCodeOutput struct {
	SessionToken string
	NextStep     domain.NextStep
	CodeType     string
	Timeout      int
}

type VerifyCodeInput struct {
	SessionToken string
	Code         string
}

func (i VerifyCodeInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" || strings.TrimSpace(i.Code) == "" {
		return ErrEmptyInput
	}
	return nil
}

type VerifyCodeOutput struct {
	NextStep domain.NextStep
	Me       *domain.Me
}

type VerifyPasswordInput struct {
	SessionToken string
	Password     string
}

func (i VerifyPasswordInput) Validate() error {
	if strings.TrimSpace(i.SessionToken) == "" || strings.TrimSpace(i.Password) == "" {
		return ErrEmptyInput
	}
	return nil
}

type VerifyPasswordOutput struct {
	NextStep domain.NextStep
	Me       domain.Me
}

type SessionInput struct {
	SessionToken string
}

type SessionOutput struct {
	Me domain.Me
}

type LogoutInput struct {
	SessionToken string
}
