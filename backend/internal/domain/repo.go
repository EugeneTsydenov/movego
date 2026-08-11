package domain

import "context"

type UserRepo interface {
	Save(ctx context.Context, user *User) error
}

type CredentialRepo interface {
	Save(ctx context.Context, cred *Credential) error
}

type SessionRepo interface {
	Save(ctx context.Context, session *Session) error
}

type Repos interface {
	Users() UserRepo
	Credentials() CredentialRepo
	Sessions() SessionRepo
}
