package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
)

func TestRegisterUseCase_Success(t *testing.T) {
	repos := newFakeRepositories()
	hasher := newFakePasswordHasher()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos}, hasher)

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "player@example.com",
		Password:    "strongpassword123",
		DisplayName: "Player One",
	})
	require.NoError(t, err)

	credRepo := repos.Credentials.(*fakeCredentialRepo)
	require.Len(t, credRepo.saved, 1, "expected exactly one credential saved")

	var accountID uuid.UUID
	var cred *identity.Credential
	for id, c := range credRepo.saved {
		accountID = id
		cred = c
	}

	accRepo := repos.Accounts.(*fakeAccountRepo)
	assert.Contains(t, accRepo.saved, accountID, "expected account to be saved")

	profRepo := repos.Profiles.(*fakeProfileRepo)
	prof, ok := profRepo.saved[accountID]
	require.True(t, ok, "expected profile to be saved")
	assert.Equal(t, "Player One", prof.DisplayName().String())
	assert.NotEmpty(t, prof.Tag().String(), "expected a generated tag")
	assert.Equal(t, 1, hasher.hashCalls)

	authzRepo := repos.Authorizations.(*fakeAuthorizationRepo)
	authz, ok := authzRepo.saved[accountID]
	require.True(t, ok, "expected authorization to be saved")
	assert.True(t, authz.HasRole(authorization.RolePlayer), "expected default role player")
	require.NotNil(t, cred)
	assert.Equal(t, "hashed:strongpassword123", cred.PasswordHash())
}

func TestRegisterUseCase_EmailAlreadyTaken(t *testing.T) {
	repos := newFakeRepositories()
	repos.Credentials.(*fakeCredentialRepo).existingEmails["taken@example.com"] = true

	hasher := newFakePasswordHasher()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos}, hasher)

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "taken@example.com",
		Password:    "strongpassword123",
		DisplayName: "Someone",
	})

	assert.ErrorIs(t, err, ErrEmailAlreadyTaken)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"no account should be saved when email is already taken")
	assert.Equal(t, 0, hasher.hashCalls)
}

func TestRegisterUseCase_InvalidEmail_NoSideEffects(t *testing.T) {
	repos := newFakeRepositories()
	hasher := newFakePasswordHasher()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos}, hasher)

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "not-an-email",
		Password:    "strongpassword123",
		DisplayName: "Someone",
	})

	require.Error(t, err)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"email validation must happen before any persistence")
	assert.Equal(t, 0, hasher.hashCalls)
}

func TestRegisterUseCase_WeakPassword_NoSideEffects(t *testing.T) {
	repos := newFakeRepositories()
	hasher := newFakePasswordHasher()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos}, hasher)

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "player@example.com",
		Password:    "short",
		DisplayName: "Someone",
	})

	require.Error(t, err)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"nothing should persist when credential creation fails partway through")
	assert.Equal(t, 0, hasher.hashCalls)
}

func TestRegisterUseCase_TagGenerationFailsAfterMaxAttempts(t *testing.T) {
	repos := newFakeRepositories()
	repos.Profiles.(*fakeProfileRepo).forceTagUsed = true

	hasher := newFakePasswordHasher()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos}, hasher)

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "collision@example.com",
		Password:    "strongpassword123",
		DisplayName: "Someone",
	})

	assert.ErrorIs(t, err, ErrTagGenerationFailed)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"no account should remain saved when tag generation exhausts all attempts")
	assert.Equal(t, 1, hasher.hashCalls)
}

func TestRegisterUseCase_GeneratedTagIsUnique(t *testing.T) {
	repos := newFakeRepositories()
	hasher := newFakePasswordHasher()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos}, hasher)

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "first@example.com",
		Password:    "strongpassword123",
		DisplayName: "First",
	})
	require.NoError(t, err)

	err = uc.Execute(context.Background(), RegisterCommand{
		Email:       "second@example.com",
		Password:    "strongpassword123",
		DisplayName: "Second",
	})
	require.NoError(t, err)

	profRepo := repos.Profiles.(*fakeProfileRepo)
	require.Len(t, profRepo.saved, 2)
	assert.Equal(t, 2, hasher.hashCalls)

	tags := make(map[string]bool)
	for _, p := range profRepo.saved {
		assert.False(t, tags[p.Tag().String()], "tags must be unique across registrations")
		tags[p.Tag().String()] = true
	}
}
