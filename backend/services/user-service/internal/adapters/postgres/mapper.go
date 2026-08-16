package postgres

import (
	"fmt"
	"time"
	"user/internal/adapters/postgres/sqlc"
	"user/internal/domain"

	"github.com/jackc/pgx/v5/pgtype"
)

func toPgTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

func toPgText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func toPgTimestampzPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{
		Time:  *t,
		Valid: true,
	}
}

func fromPgTextPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func fromPgTimestampzPtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

func toSaveCredentialsParams(cred *domain.Credential) sqlc.SaveCredentialParams {
	return sqlc.SaveCredentialParams{
		ID:           cred.ID(),
		UserID:       cred.UserID(),
		PasswordHash: toPgTextPtr(cred.PasswordHash()),
		Provider:     cred.Provider().String(),
		ProviderKey:  toPgTextPtr(cred.ProviderKey()),
	}
}

func toSaveSessionsParams(session *domain.Session) sqlc.SaveSessionParams {
	return sqlc.SaveSessionParams{
		ID:           session.ID(),
		UserID:       session.UserID(),
		SecretHash:   session.SecretHash(),
		UserAgent:    session.UserAgent(),
		ClientIp:     session.ClientIP(),
		LastActiveAt: session.LastActiveAt(),
		ExpiresAt:    session.ExpiresAt(),
		CreatedAt:    session.CreatedAt(),
	}
}

func toSaveUsersParams(user *domain.User) sqlc.SaveUserParams {
	return sqlc.SaveUserParams{
		ID:          user.ID(),
		Email:       user.Email().String(),
		Tag:         user.Tag().String(),
		DisplayName: user.DisplayName().String(),
		Role:        user.Role().String(),
		CreatedAt:   user.CreatedAt(),
		UpdatedAt:   user.UpdatedAt(),
		DeletedAt:   toPgTimestampzPtr(user.DeletedAt()),
	}
}

func toDomainSession(session sqlc.Sessions) *domain.Session {
	return domain.RestoreSession(
		session.ID,
		session.UserID,
		session.SecretHash,
		session.UserAgent,
		session.ClientIp,
		session.LastActiveAt,
		session.ExpiresAt,
		session.CreatedAt,
	)
}

func toDomainUser(row sqlc.Users) (*domain.User, error) {
	email, err := domain.NewEmail(row.Email)
	if err != nil {
		return nil, fmt.Errorf("map email: %v", err)
	}

	tag, err := domain.NewTag(row.Tag)
	if err != nil {
		return nil, fmt.Errorf("map tag: %v", err)
	}

	displayName, err := domain.NewDisplayName(row.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("map display name: %v", err)
	}

	role, err := domain.NewRole(row.Role)
	if err != nil {
		return nil, fmt.Errorf("map role: %v", err)
	}

	return domain.RestoreUser(
		row.ID,
		email,
		tag,
		displayName,
		role,
		row.CreatedAt,
		row.UpdatedAt,
		fromPgTimestampzPtr(row.DeletedAt),
	), nil
}

func toDomainCredential(row sqlc.Credentials) (*domain.Credential, error) {
	provider, err := domain.NewProvider(row.Provider)
	if err != nil {
		return nil, fmt.Errorf("map provider: %v", err)
	}

	return domain.RestoreCredential(
		row.ID,
		row.UserID,
		provider,
		fromPgTextPtr(row.PasswordHash),
		fromPgTextPtr(row.ProviderKey),
	), nil
}

func toFindForAuthParams(email domain.Email, provider domain.Provider) sqlc.FindForAuthParams {
	return sqlc.FindForAuthParams{
		Email:    email.String(),
		Provider: provider.String(),
	}
}

func toDomainFindForAuthRow(row sqlc.FindForAuthRow) (*domain.User, *domain.Credential, error) {
	user, err := toDomainUser(row.Users)
	if err != nil {
		return nil, nil, err
	}

	credential, err := toDomainCredential(row.Credentials)
	if err != nil {
		return nil, nil, err
	}

	return user, credential, nil
}

func toDomainFindValidRow(row sqlc.FindValidRow) (*domain.Session, *domain.User, error) {
	user, err := toDomainUser(row.Users)
	if err != nil {
		return nil, nil, err
	}
	session := toDomainSession(row.Sessions)
	return session, user, nil
}

func toDomainSessionList(rows []sqlc.Sessions) []*domain.Session {
	sessions := make([]*domain.Session, len(rows))
	for i, row := range rows {
		sessions[i] = toDomainSession(row)
	}
	return sessions
}
