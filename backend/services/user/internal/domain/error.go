package domain

import "fmt"

type ErrorKind int

const (
	KindUnknown ErrorKind = iota
	KindNotFound
	KindAlreadyExists
	KindInvalidInput
	KindPermissionDenied
	KindConflict
	KindUnauthenticated
	KindInternal
)

type DomainError struct {
	Kind    ErrorKind
	Code    string
	Message string
	Details map[string]any
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("%s", e.Message)
}

func (e *DomainError) Unwrap() error { return e.Err }

func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	return ok && e.Code == t.Code
}

func NewNotFound(code, msg string, details map[string]any) *DomainError {
	return &DomainError{Kind: KindNotFound, Code: code, Message: msg, Details: details}
}

func NewConflict(code, msg string, err error) *DomainError {
	return &DomainError{Kind: KindConflict, Code: code, Message: msg, Err: err}
}

func NewInvalidInput(code, msg string) *DomainError {
	return &DomainError{Kind: KindInvalidInput, Code: code, Message: msg}
}

func NewUnauthenticated(code, msg string) *DomainError {
	return &DomainError{Kind: KindUnauthenticated, Code: code, Message: msg}
}
