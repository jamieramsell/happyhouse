package users

import (
	"time"

	"github.com/google/uuid"
)

// A user is a registered account
type User struct {
	Id           uuid.UUID
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
