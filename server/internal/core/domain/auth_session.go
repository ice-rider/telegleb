package domain

import "time"

type SessionStatus string

const (
	StatusAwaitingCode     SessionStatus = "AWAITING_CODE"
	StatusAwaitingPassword SessionStatus = "AWAITING_PASSWORD"
	StatusAuthorized       SessionStatus = "AUTHORIZED"
)

type NextStep string

const (
	NextStepCode     NextStep = "code"
	NextStepPassword NextStep = "password"
	NextStepDone     NextStep = "done"
)

type AuthSession struct {
	SessionToken        string
	PhoneNumber         string
	TelegramCodeHash    string
	Status              SessionStatus
	TelegramSessionData []byte
	ExpiresAt           time.Time
}
