package application

import (
	"time"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID          uuid.UUID
	Tag         string
	Email       string
	DisplayName string
	Role        string
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

type SessionDTO struct {
	ID           uuid.UUID
	UserAgent    string
	ClientIP     string
	IsCurrent    bool
	LastActiveAt time.Time
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type SignUpInput struct {
	Email     string
	Password  string
	UserAgent string
	ClientIP  string
}

type SignUpOutput struct {
	User         UserDTO
	AccessToken  string
	RefreshToken string
}

type SignInInput struct {
	Email     string
	Password  string
	UserAgent string
	ClientIP  string
}

type SignInOutput struct {
	User         UserDTO
	AccessToken  string
	RefreshToken string
}

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

type SignOutInput struct {
	RefreshToken string
}

type SignOutOutput struct{}
