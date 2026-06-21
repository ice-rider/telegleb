package domain

import "time"

type SessionStatus string

const (
	AWAITING_PHONE    SessionStatus = "AWAITING_PHONE"
	AWAITING_CODE     SessionStatus = "AWAITING_CODE"
	AWAITING_PASSWORD SessionStatus = "AWAITING_PASSWORD"
	AUTHORIZED        SessionStatus = "AUTHORIZED"
)

type NextStep string

const (
	NextStepAwaitingPassword NextStep = "AWAITING_PASSWORD"
	NextStepAuthorized       NextStep = "AUTHORIZED"
)

type AuthSession struct {
	SessionToken        string
	PhoneNumber         string
	TelegramCodeHash    string
	Status              SessionStatus
	TelegramSessionData []byte
	ExpiresAt           time.Time
}
