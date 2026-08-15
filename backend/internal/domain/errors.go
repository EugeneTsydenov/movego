package domain

import (
	"errors"
)

// Kind Errors
var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrValidation     = errors.New("validation failed")
	ErrAuthentication = errors.New("authentication failed")
)

type domainError struct {
	kind    error
	message string
}

func (e *domainError) Error() string {
	return e.message
}

func (e *domainError) Unwrap() error {
	return e.kind
}

func newErr(kind error, message string) error {
	return &domainError{
		kind:    kind,
		message: message,
	}
}

var (
	// Not Found
	ErrUserNotFound       = newErr(ErrNotFound, "user not found")
	ErrSessionNotFound    = newErr(ErrNotFound, "session not found")
	ErrCredentialNotFound = newErr(ErrNotFound, "credential not found")

	// Already Exists
	ErrEmailTaken            = newErr(ErrAlreadyExists, "email is already taken")
	ErrTagTaken              = newErr(ErrAlreadyExists, "tag is already taken")
	ErrProviderAlreadyLinked = newErr(ErrAlreadyExists, "provider is already linked")
	ErrProviderKeyTaken      = newErr(ErrAlreadyExists, "social account is already linked")

	// Validation
	ErrInvalidDisplayName = newErr(ErrValidation, "invalid display name")
	ErrInvalidEmail       = newErr(ErrValidation, "invalid email")
	ErrInvalidTag         = newErr(ErrValidation, "invalid tag")
	ErrWeakPassword       = newErr(ErrValidation, "password is too weak")
	ErrInvalidRole        = newErr(ErrValidation, "invalid role")
	ErrInvalidProvider    = newErr(ErrValidation, "invalid provider")

	// Authentication
	ErrInvalidCredentials = newErr(ErrAuthentication, "invalid credentials")
)
