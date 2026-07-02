package authorization

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrCannotDemoteLastAdmin = errors.New("authorization: cannot demote the last admin")

type Role string

const (
	RolePlayer Role = "player"
	RoleAdmin  Role = "admin"
)

func (r Role) String() string {
	return string(r)
}

type Authorization struct {
	accountID uuid.UUID
	role      Role
	createdAt time.Time
	updatedAt time.Time
}

func New(accountID uuid.UUID) *Authorization {
	now := time.Now().UTC()
	return &Authorization{
		accountID: accountID,
		role:      RolePlayer,
		createdAt: now,
		updatedAt: now,
	}
}

func (a *Authorization) AccountID() uuid.UUID { return a.accountID }
func (a *Authorization) Role() Role           { return a.role }
func (a *Authorization) CreatedAt() time.Time { return a.createdAt }
func (a *Authorization) UpdatedAt() time.Time { return a.updatedAt }

func (a *Authorization) PromoteToAdmin() {
	a.role = RoleAdmin
	a.updatedAt = time.Now().UTC()
}

func (a *Authorization) DemoteToPlayer() {
	a.role = RolePlayer
	a.updatedAt = time.Now().UTC()
}

func (a *Authorization) HasRole(r Role) bool {
	return a.role == r
}

func (a *Authorization) IsAdmin() bool {
	return a.role == RoleAdmin
}
