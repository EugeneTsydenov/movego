package authorization

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	accountID := uuid.New()

	auth := New(accountID)

	assert.NotNil(t, auth)
	assert.Equal(t, accountID, auth.AccountID())
	assert.Equal(t, RolePlayer, auth.Role())
	assert.False(t, auth.CreatedAt().IsZero())
	assert.False(t, auth.UpdatedAt().IsZero())
	assert.Equal(t, auth.CreatedAt(), auth.UpdatedAt())
}

func TestPromoteToAdmin(t *testing.T) {
	auth := New(uuid.New())

	auth.PromoteToAdmin()

	assert.Equal(t, RoleAdmin, auth.Role())
	assert.True(t, auth.IsAdmin())
	assert.True(t, auth.HasRole(RoleAdmin))
	assert.False(t, auth.HasRole(RolePlayer))
	assert.True(t, auth.UpdatedAt().After(auth.CreatedAt()))
}

func TestDemoteToPlayer(t *testing.T) {
	auth := New(uuid.New())

	auth.PromoteToAdmin()
	auth.DemoteToPlayer()

	assert.Equal(t, RolePlayer, auth.Role())
	assert.False(t, auth.IsAdmin())
	assert.True(t, auth.HasRole(RolePlayer))
	assert.False(t, auth.HasRole(RoleAdmin))
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(*Authorization)
		role     Role
		expected bool
	}{
		{
			name:     "player has player role",
			prepare:  func(a *Authorization) {},
			role:     RolePlayer,
			expected: true,
		},
		{
			name:     "player does not have admin role",
			prepare:  func(a *Authorization) {},
			role:     RoleAdmin,
			expected: false,
		},
		{
			name: "admin has admin role",
			prepare: func(a *Authorization) {
				a.PromoteToAdmin()
			},
			role:     RoleAdmin,
			expected: true,
		},
		{
			name: "admin does not have player role",
			prepare: func(a *Authorization) {
				a.PromoteToAdmin()
			},
			role:     RolePlayer,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := New(uuid.New())
			tt.prepare(auth)

			assert.Equal(t, tt.expected, auth.HasRole(tt.role))
		})
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(*Authorization)
		expected bool
	}{
		{
			name:     "player",
			prepare:  func(a *Authorization) {},
			expected: false,
		},
		{
			name: "admin",
			prepare: func(a *Authorization) {
				a.PromoteToAdmin()
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := New(uuid.New())
			tt.prepare(auth)

			assert.Equal(t, tt.expected, auth.IsAdmin())
		})
	}
}
