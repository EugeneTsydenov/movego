package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	acc := New()

	assert.NotNil(t, acc)
	assert.Equal(t, Active, acc.Status())
	assert.False(t, acc.ID().String() == "")
	assert.False(t, acc.CreatedAt().IsZero())
	assert.False(t, acc.UpdatedAt().IsZero())
	assert.Equal(t, acc.CreatedAt(), acc.UpdatedAt())
}

func TestSuspend(t *testing.T) {
	tests := []struct {
		name        string
		prepare     func(*Account)
		expectedErr error
		expected    Status
	}{
		{
			name:        "active to suspended",
			prepare:     func(a *Account) {},
			expectedErr: nil,
			expected:    Suspended,
		},
		{
			name: "deleted account",
			prepare: func(a *Account) {
				a.Delete()
			},
			expectedErr: ErrAlreadyDeleted,
			expected:    Deleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := New()
			tt.prepare(acc)

			err := acc.Suspend()

			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expected, acc.Status())
		})
	}
}

func TestActivateDeletedAccount(t *testing.T) {
	acc := New()
	acc.Delete()

	err := acc.Activate()

	assert.ErrorIs(t, err, ErrAlreadyDeleted)
	assert.Equal(t, Deleted, acc.Status())
}

func TestActivate(t *testing.T) {
	tests := []struct {
		name        string
		prepare     func(*Account)
		expectedErr error
		expected    Status
	}{
		{
			name:        "to active",
			prepare:     func(a *Account) {},
			expectedErr: nil,
			expected:    Active,
		},
		{
			name: "deleted account",
			prepare: func(a *Account) {
				a.Delete()
			},
			expectedErr: ErrAlreadyDeleted,
			expected:    Deleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := New()
			tt.prepare(acc)

			err := acc.Activate()

			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expected, acc.Status())
		})
	}
}

func TestDelete(t *testing.T) {
	acc := New()

	acc.Delete()

	assert.Equal(t, Deleted, acc.Status())
}

func TestIsActive(t *testing.T) {
	acc := New()

	assert.True(t, acc.IsActive())

	_ = acc.Suspend()
	assert.False(t, acc.IsActive())

	acc.Delete()
	assert.False(t, acc.IsActive())
}
