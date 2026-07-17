package account

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAlreadyDeleted  = errors.New("account: already deleted")
	ErrAccountNotFound = errors.New("account: not found")
)

type Status int

const (
	Active Status = iota
	Suspended
	Deleted
)

func (s Status) String() string {
	switch s {
	case Active:
		return "active"
	case Suspended:
		return "suspended"
	case Deleted:
		return "deleted"
	default:
		return "unknown"
	}
}

type Account struct {
	id        uuid.UUID
	status    Status
	updatedAt time.Time
	createdAt time.Time
}

func New() *Account {
	now := time.Now().UTC()
	return &Account{
		id:        uuid.Must(uuid.NewV7()),
		status:    Active,
		updatedAt: now,
		createdAt: now,
	}
}

func (a *Account) ID() uuid.UUID        { return a.id }
func (a *Account) Status() Status       { return a.status }
func (a *Account) CreatedAt() time.Time { return a.createdAt }
func (a *Account) UpdatedAt() time.Time { return a.updatedAt }

func (a *Account) Suspend() error {
	if a.status == Deleted {
		return ErrAlreadyDeleted
	}
	a.status = Suspended
	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Activate() error {
	if a.status == Deleted {
		return ErrAlreadyDeleted
	}
	a.status = Active
	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Delete() {
	a.status = Deleted
	a.updatedAt = time.Now().UTC()
}

func (a *Account) IsActive() bool {
	return a.status == Active
}
