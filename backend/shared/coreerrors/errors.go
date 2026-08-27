package coreerrors

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrValidation       = errors.New("validation failed")
	ErrAuthentication   = errors.New("authentication failed")
	ErrPermissionDenied = errors.New("permission denied")
)

type DomainError struct {
	kind    error
	message string
}

func (e *DomainError) Error() string {
	return e.message
}

func (e *DomainError) Unwrap() error {
	return e.kind
}

func New(kind error, message string) error {
	return &DomainError{
		kind:    kind,
		message: message,
	}
}
