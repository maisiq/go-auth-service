package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	UserRole    Role = "user"
	PremiumRole Role = "premium"
	AdminRole   Role = "admin"
)

type User struct {
	ID             uuid.UUID
	Email          string
	Role           Role
	HashedPassword string `db:"hashed_password"`

	LastLoogedAt time.Time `db:"last_logged_at"`
	CreatedAT    time.Time `db:"created_at"`

	SocialAccount  bool // Tells whether this user was created via one of OAuth providers
	SocialID       string
	SocialProvider string
}

// UserLog is a record for user's each logging try
type UserLog struct {
	ID        uuid.UUID `json:"id"`
	UserEmail string    `json:"user_email"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	LoggedAt  time.Time `json:"logged_at"`
}
