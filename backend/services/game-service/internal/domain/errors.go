package domain

import (
	"errors"
)

// Kind Errors
var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("not found")
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
	// Validation
	ErrInvalidTimeControl = newErr(ErrValidation, "invalid time control")
	ErrGameNotFound       = newErr(ErrNotFound, "game not found")
)
