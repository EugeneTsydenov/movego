package domain

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

const suffixChars = "abcdefghijklmnopqrstuvwxyz0123456789"

var chessPieces = []string{
	"pawn", "knight", "bishop", "rook", "queen", "king",
}

type User struct {
	id          uuid.UUID
	email       Email
	tag         Tag
	displayName DisplayName
	role        Role
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

func NewUser(email Email, tag Tag, displayName DisplayName, role Role) *User {
	now := time.Now()
	return &User{
		id:          uuid.Must(uuid.NewV7()),
		email:       email,
		tag:         tag,
		displayName: displayName,
		role:        role,
		createdAt:   now,
		updatedAt:   now,
	}
}

func NewTemporaryUser(email Email, role Role) *User {
	return NewUser(email, generateChessTagCandidate(), generateChessDisplayNameCandidate(email), role)
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Tag() Tag {
	return u.tag
}

func (u *User) DisplayName() DisplayName {
	return u.displayName
}

func (u *User) Role() Role {
	return u.role
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) DeletedAt() *time.Time {
	return u.deletedAt
}

func secureRandN(n int) int {
	v, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(v.Int64())
}

func randomSuffix(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = suffixChars[secureRandN(len(suffixChars))]
	}
	return string(b)
}

func generateChessTagCandidate() Tag {
	piece := chessPieces[secureRandN(len(chessPieces))]
	suffix := randomSuffix(6)
	tag, _ := NewTag(piece + suffix) // "knight7x9k2m"
	return tag
}

func generateChessDisplayNameCandidate(email Email) DisplayName {
	raw := email.String() // "oleg@gmail.com"
	if at := strings.Index(raw, "@"); at > 0 {
		raw = raw[:at] // "oleg"
	}
	name, _ := NewDisplayName(raw)
	return name
}
