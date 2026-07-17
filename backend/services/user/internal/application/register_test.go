package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/movego/services/user/internal/domain/account"
	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
	"github.com/movego/services/user/internal/domain/profile"
)

// FAKES
type fakeAccountRepo struct {
	saved map[uuid.UUID]*account.Account
}

func newFakeAccountRepo() *fakeAccountRepo {
	return &fakeAccountRepo{saved: map[uuid.UUID]*account.Account{}}
}

func (r *fakeAccountRepo) Save(ctx context.Context, a *account.Account) error {
	r.saved[a.ID()] = a
	return nil
}

func (r *fakeAccountRepo) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	a, ok := r.saved[id]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	return a, nil
}

type fakeCredentialRepo struct {
	saved          map[uuid.UUID]*identity.Credential
	existingEmails map[string]bool
}

func newFakeCredentialRepo() *fakeCredentialRepo {
	return &fakeCredentialRepo{
		saved:          map[uuid.UUID]*identity.Credential{},
		existingEmails: map[string]bool{},
	}
}

func (r *fakeCredentialRepo) Save(ctx context.Context, c *identity.Credential) error {
	r.saved[c.ID()] = c
	return nil
}

func (r *fakeCredentialRepo) FindByID(ctx context.Context, id uuid.UUID) (*identity.Credential, error) {
	c, ok := r.saved[id]
	if !ok {
		return nil, identity.ErrCredentialNotFound
	}
	return c, nil
}

func (r *fakeCredentialRepo) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*identity.Credential, error) {
	for _, c := range r.saved {
		if c.AccountID() == accountID {
			return c, nil
		}
	}
	return nil, identity.ErrCredentialNotFound
}

func (r *fakeCredentialRepo) FindByEmail(ctx context.Context, email identity.Email) (*identity.Credential, error) {
	for _, c := range r.saved {
		if c.Email() == email.String() {
			return c, nil
		}
	}
	return nil, identity.ErrCredentialNotFound
}

func (r *fakeCredentialRepo) ExistsByEmail(ctx context.Context, email identity.Email) (bool, error) {
	return r.existingEmails[email.String()], nil
}

type fakeProfileRepo struct {
	saved        map[uuid.UUID]*profile.Profile
	forceTagUsed bool
}

func newFakeProfileRepo() *fakeProfileRepo {
	return &fakeProfileRepo{saved: map[uuid.UUID]*profile.Profile{}}
}

func (r *fakeProfileRepo) Save(ctx context.Context, p *profile.Profile) error {
	r.saved[p.AccountID()] = p
	return nil
}

func (r *fakeProfileRepo) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*profile.Profile, error) {
	p, ok := r.saved[accountID]
	if !ok {
		return nil, profile.ErrProfileNotFound
	}
	return p, nil
}

func (r *fakeProfileRepo) FindByTag(ctx context.Context, tag profile.Tag) (*profile.Profile, error) {
	for _, p := range r.saved {
		if p.Tag() == tag {
			return p, nil
		}
	}
	return nil, profile.ErrProfileNotFound
}

func (r *fakeProfileRepo) ExistsByTag(ctx context.Context, tag profile.Tag) (bool, error) {
	if r.forceTagUsed {
		return true, nil
	}
	for _, p := range r.saved {
		if p.Tag() == tag {
			return true, nil
		}
	}
	return false, nil
}

type fakeAuthorizationRepo struct {
	saved map[uuid.UUID]*authorization.Authorization
}

func newFakeAuthorizationRepo() *fakeAuthorizationRepo {
	return &fakeAuthorizationRepo{saved: map[uuid.UUID]*authorization.Authorization{}}
}

func (r *fakeAuthorizationRepo) Save(ctx context.Context, a *authorization.Authorization) error {
	r.saved[a.AccountID()] = a
	return nil
}

func (r *fakeAuthorizationRepo) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*authorization.Authorization, error) {
	a, ok := r.saved[accountID]
	if !ok {
		return nil, authorization.ErrAuthorizationNotFound
	}
	return a, nil
}

func (r *fakeAuthorizationRepo) CountByRole(ctx context.Context, role authorization.Role) (int, error) {
	count := 0
	for _, a := range r.saved {
		if a.Role() == role {
			count++
		}
	}
	return count, nil
}

type fakeUOW struct {
	repos *Repositories
}

func (u *fakeUOW) Do(ctx context.Context, fn func(repos *Repositories) error) error {
	return fn(u.repos)
}

func newFakeRepositories() *Repositories {
	return &Repositories{
		Accounts:       newFakeAccountRepo(),
		Credentials:    newFakeCredentialRepo(),
		Profiles:       newFakeProfileRepo(),
		Authorizations: newFakeAuthorizationRepo(),
	}
}

// TESTS
func TestRegisterUseCase_Success(t *testing.T) {
	repos := newFakeRepositories()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos})

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "player@example.com",
		Password:    "strongpassword123",
		DisplayName: "Player One",
	})
	require.NoError(t, err)

	credRepo := repos.Credentials.(*fakeCredentialRepo)
	require.Len(t, credRepo.saved, 1, "expected exactly one credential saved")

	var accountID uuid.UUID
	for _, c := range credRepo.saved {
		accountID = c.AccountID()
	}

	accRepo := repos.Accounts.(*fakeAccountRepo)
	assert.Contains(t, accRepo.saved, accountID, "expected account to be saved")

	profRepo := repos.Profiles.(*fakeProfileRepo)
	prof, ok := profRepo.saved[accountID]
	require.True(t, ok, "expected profile to be saved")
	assert.Equal(t, "Player One", prof.DisplayName().String())
	assert.NotEmpty(t, prof.Tag().String(), "expected a generated tag")

	authzRepo := repos.Authorizations.(*fakeAuthorizationRepo)
	authz, ok := authzRepo.saved[accountID]
	require.True(t, ok, "expected authorization to be saved")
	assert.True(t, authz.HasRole(authorization.RolePlayer), "expected default role player")
}

func TestRegisterUseCase_EmailAlreadyTaken(t *testing.T) {
	repos := newFakeRepositories()
	repos.Credentials.(*fakeCredentialRepo).existingEmails["taken@example.com"] = true

	uc := NewRegisterUseCase(&fakeUOW{repos: repos})

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "taken@example.com",
		Password:    "strongpassword123",
		DisplayName: "Someone",
	})

	assert.ErrorIs(t, err, ErrEmailAlreadyTaken)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"no account should be saved when email is already taken")
}

func TestRegisterUseCase_InvalidEmail_NoSideEffects(t *testing.T) {
	repos := newFakeRepositories()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos})

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "not-an-email",
		Password:    "strongpassword123",
		DisplayName: "Someone",
	})

	require.Error(t, err)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"email validation must happen before any persistence")
}

func TestRegisterUseCase_WeakPassword_NoSideEffects(t *testing.T) {
	repos := newFakeRepositories()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos})

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "player@example.com",
		Password:    "short",
		DisplayName: "Someone",
	})

	require.Error(t, err)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"nothing should persist when credential creation fails partway through")
}

func TestRegisterUseCase_TagGenerationFailsAfterMaxAttempts(t *testing.T) {
	repos := newFakeRepositories()
	repos.Profiles.(*fakeProfileRepo).forceTagUsed = true

	uc := NewRegisterUseCase(&fakeUOW{repos: repos})

	err := uc.Execute(context.Background(), RegisterCommand{
		Email:       "collision@example.com",
		Password:    "strongpassword123",
		DisplayName: "Someone",
	})

	assert.ErrorIs(t, err, ErrTagGenerationFailed)
	assert.Empty(t, repos.Accounts.(*fakeAccountRepo).saved,
		"no account should remain saved when tag generation exhausts all attempts")
}

func TestRegisterUseCase_GeneratedTagIsUnique(t *testing.T) {
	repos := newFakeRepositories()
	uc := NewRegisterUseCase(&fakeUOW{repos: repos})

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

	tags := make(map[string]bool)
	for _, p := range profRepo.saved {
		assert.False(t, tags[p.Tag().String()], "tags must be unique across registrations")
		tags[p.Tag().String()] = true
	}
}
