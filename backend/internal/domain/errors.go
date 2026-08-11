package domain

import (
	"errors"
	"fmt"
)

// Kind Errors
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrValidation    = errors.New("validation failed")
)

var (
	// Not Found
	ErrUserNotFound       = fmt.Errorf("user: %w", ErrNotFound)
	ErrSessionNotFound    = fmt.Errorf("session: %w", ErrNotFound)
	ErrCredentialNotFound = fmt.Errorf("credential: %w", ErrNotFound)

	// Already Exists
	ErrEmailTaken            = fmt.Errorf("email: %w", ErrAlreadyExists)
	ErrTagTaken              = fmt.Errorf("tag: %w", ErrAlreadyExists)
	ErrProviderAlreadyLinked = fmt.Errorf("provider: %w", ErrAlreadyExists)
	ErrProviderKeyTaken      = fmt.Errorf("social account: %w", ErrAlreadyExists)

	// Validation
	ErrInvalidDisplayName = fmt.Errorf("display name: %w", ErrValidation)
	ErrInvalidEmail       = fmt.Errorf("email: %w", ErrValidation)
	ErrInvalidTag         = fmt.Errorf("tag: %w", ErrValidation)
	ErrWeakPassword       = fmt.Errorf("password: %w", ErrValidation)
	ErrInvalidRole        = fmt.Errorf("role: %w", ErrValidation)
	ErrInvalidProvider    = fmt.Errorf("provider: %w", ErrValidation)
)
