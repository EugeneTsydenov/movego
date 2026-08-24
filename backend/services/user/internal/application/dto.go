package application

import (
	"time"
	"user/internal/domain"
)

type UserDTO struct {
	ID          domain.UserID
	Tag         domain.Tag
	Email       domain.Email
	DisplayName domain.DisplayName
	Role        domain.Role
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

type SessionDTO struct {
	ID           domain.SessionID
	UserAgent    string
	ClientIP     string
	IsCurrent    bool
	LastActiveAt time.Time
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type SignUpInput struct {
	Email     domain.Email
	Password  domain.PlainPassword
	UserAgent string
	ClientIP  string
}

type SignUpOutput struct {
	User         UserDTO
	AccessToken  string
	RefreshToken string
}

type SignInInput struct {
	Email     domain.Email
	Password  domain.PlainPassword
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
