package domain

import "github.com/google/uuid"

type User struct {
	ID         uuid.UUID
	TelegramID int64
	FirstName  string
	LastName   string
	Username   string
	Phone      string
	IsBot      bool
}
