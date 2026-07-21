package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/movego/services/user/internal/domain/account"
	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
	"github.com/movego/services/user/internal/domain/profile"
)

type fakeAccountRepo struct {
	saved   map[uuid.UUID]*account.Account
	saveErr error
	findErr error
}

func newFakeAccountRepo() *fakeAccountRepo {
	return &fakeAccountRepo{saved: map[uuid.UUID]*account.Account{}}
}

func (r *fakeAccountRepo) Save(ctx context.Context, a *account.Account) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved[a.ID()] = a
	return nil
}

func (r *fakeAccountRepo) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	a, ok := r.saved[id]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	return a, nil
}

type fakeCredentialRepo struct {
	saved            map[uuid.UUID]*identity.Credential
	existingEmails   map[string]bool
	saveErr          error
	findByIDErr      error
	findByAccountErr error
	findByEmailErr   error
	existsByEmailErr error
	findByEmailCalls int
}

func newFakeCredentialRepo() *fakeCredentialRepo {
	return &fakeCredentialRepo{
		saved:          map[uuid.UUID]*identity.Credential{},
		existingEmails: map[string]bool{},
	}
}

func (r *fakeCredentialRepo) Save(ctx context.Context, c *identity.Credential) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved[c.ID()] = c
	return nil
}

func (r *fakeCredentialRepo) FindByID(ctx context.Context, id uuid.UUID) (*identity.Credential, error) {
	if r.findByIDErr != nil {
		return nil, r.findByIDErr
	}
	c, ok := r.saved[id]
	if !ok {
		return nil, identity.ErrCredentialNotFound
	}
	return c, nil
}

func (r *fakeCredentialRepo) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*identity.Credential, error) {
	if r.findByAccountErr != nil {
		return nil, r.findByAccountErr
	}
	for _, c := range r.saved {
		if c.AccountID() == accountID {
			return c, nil
		}
	}
	return nil, identity.ErrCredentialNotFound
}

func (r *fakeCredentialRepo) FindByEmail(ctx context.Context, email identity.Email) (*identity.Credential, error) {
	r.findByEmailCalls++
	if r.findByEmailErr != nil {
		return nil, r.findByEmailErr
	}
	for _, c := range r.saved {
		if c.Email() == email.String() {
			return c, nil
		}
	}
	return nil, identity.ErrCredentialNotFound
}

func (r *fakeCredentialRepo) ExistsByEmail(ctx context.Context, email identity.Email) (bool, error) {
	if r.existsByEmailErr != nil {
		return false, r.existsByEmailErr
	}
	return r.existingEmails[email.String()], nil
}

type fakeProfileRepo struct {
	saved        map[uuid.UUID]*profile.Profile
	forceTagUsed bool
	saveErr      error
	findErr      error
	findByTagErr error
	existsErr    error
}

func newFakeProfileRepo() *fakeProfileRepo {
	return &fakeProfileRepo{saved: map[uuid.UUID]*profile.Profile{}}
}

func (r *fakeProfileRepo) Save(ctx context.Context, p *profile.Profile) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved[p.AccountID()] = p
	return nil
}

func (r *fakeProfileRepo) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*profile.Profile, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	p, ok := r.saved[accountID]
	if !ok {
		return nil, profile.ErrProfileNotFound
	}
	return p, nil
}

func (r *fakeProfileRepo) FindByTag(ctx context.Context, tag profile.Tag) (*profile.Profile, error) {
	if r.findByTagErr != nil {
		return nil, r.findByTagErr
	}
	for _, p := range r.saved {
		if p.Tag() == tag {
			return p, nil
		}
	}
	return nil, profile.ErrProfileNotFound
}

func (r *fakeProfileRepo) ExistsByTag(ctx context.Context, tag profile.Tag) (bool, error) {
	if r.existsErr != nil {
		return false, r.existsErr
	}
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
	saved   map[uuid.UUID]*authorization.Authorization
	saveErr error
	findErr error
}

func newFakeAuthorizationRepo() *fakeAuthorizationRepo {
	return &fakeAuthorizationRepo{saved: map[uuid.UUID]*authorization.Authorization{}}
}

func (r *fakeAuthorizationRepo) Save(ctx context.Context, a *authorization.Authorization) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved[a.AccountID()] = a
	return nil
}

func (r *fakeAuthorizationRepo) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*authorization.Authorization, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
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

type fakeSessionRepo struct {
	saved             map[uuid.UUID]*identity.Session
	saveErr           error
	findByTokenErr    error
	findAllByAccErr   error
	revokeByIDErr     error
	revokeAllByAccErr error
	saveCalls         int
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{saved: map[uuid.UUID]*identity.Session{}}
}

func (r *fakeSessionRepo) Save(ctx context.Context, s *identity.Session) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saveCalls++
	r.saved[s.ID()] = s
	return nil
}

func (r *fakeSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*identity.Session, error) {
	if r.findByTokenErr != nil {
		return nil, r.findByTokenErr
	}
	for _, s := range r.saved {
		if s.TokenHash() == tokenHash {
			return s, nil
		}
	}
	return nil, identity.ErrSessionNotFound
}

func (r *fakeSessionRepo) FindAllByAccountID(ctx context.Context, accountID uuid.UUID) ([]*identity.Session, error) {
	if r.findAllByAccErr != nil {
		return nil, r.findAllByAccErr
	}
	result := make([]*identity.Session, 0)
	for _, s := range r.saved {
		if s.AccountID() == accountID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *fakeSessionRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	if r.revokeByIDErr != nil {
		return r.revokeByIDErr
	}
	s, ok := r.saved[id]
	if !ok {
		return identity.ErrSessionNotFound
	}
	s.Revoke()
	return nil
}

func (r *fakeSessionRepo) RevokeAllByAccountID(ctx context.Context, accountID uuid.UUID) error {
	if r.revokeAllByAccErr != nil {
		return r.revokeAllByAccErr
	}
	found := false
	for _, s := range r.saved {
		if s.AccountID() == accountID {
			s.Revoke()
			found = true
		}
	}
	if !found {
		return identity.ErrSessionNotFound
	}
	return nil
}

func (r *fakeSessionRepo) findNonRevokedByTokenHash(tokenHash string) (*identity.Session, error) {
	for _, s := range r.saved {
		if s.TokenHash() == tokenHash && !s.IsRevoked() {
			return s, nil
		}
	}
	return nil, identity.ErrSessionNotFound
}

type fakeUOW struct {
	repos *Repositories
	doErr error
}

func (u *fakeUOW) Do(ctx context.Context, fn func(repos *Repositories) error) error {
	if u.doErr != nil {
		return u.doErr
	}
	return fn(u.repos)
}

type fakeTokenIssuer struct {
	accessToken string
	expiresAt   time.Time
	issueErr    error
	parseClaims TokenClaims
	parseErr    error
	lastClaims  TokenClaims
	issueCalls  int
}

func (i *fakeTokenIssuer) IssueAccessToken(ctx context.Context, claims TokenClaims) (string, time.Time, error) {
	if i.issueErr != nil {
		return "", time.Time{}, i.issueErr
	}
	i.lastClaims = claims
	i.issueCalls++
	if i.expiresAt.IsZero() {
		i.expiresAt = time.Now().UTC().Add(15 * time.Minute)
	}
	if i.accessToken == "" {
		i.accessToken = "fake token"
	}
	return i.accessToken, i.expiresAt, nil
}

func (i *fakeTokenIssuer) ParseAccessToken(ctx context.Context, token string) (TokenClaims, error) {
	if i.parseErr != nil {
		return TokenClaims{}, i.parseErr
	}
	return i.parseClaims, nil
}

func newFakeRepositories() *Repositories {
	return &Repositories{
		Accounts:       newFakeAccountRepo(),
		Credentials:    newFakeCredentialRepo(),
		Sessions:       newFakeSessionRepo(),
		Profiles:       newFakeProfileRepo(),
		Authorizations: newFakeAuthorizationRepo(),
	}
}
