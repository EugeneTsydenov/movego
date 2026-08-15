package postgres

import (
	"fmt"
	"movego/internal/adapters/postgres/sqlc"
	"movego/internal/domain"
	"time"

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
		ID:         session.ID(),
		UserID:     session.UserID(),
		SecretHash: session.SecretHash(),
		UserAgent:  session.UserAgent(),
		ClientIp:   session.ClientIP(),
		ExpiresAt:  session.ExpiresAt(),
		CreatedAt:  session.CreatedAt(),
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

func toFindForAuthParams(email domain.Email, provider domain.Provider) sqlc.FindForAuthParams {
	return sqlc.FindForAuthParams{
		Email:    email.String(),
		Provider: provider.String(),
	}
}

func toDomainFindForAuthRow(row sqlc.FindForAuthRow) (*domain.User, *domain.Credential, error) {
	userID := row.Users.ID
	credentialUserID := row.Credentials.UserID

	email, err := domain.NewEmail(row.Users.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("map email: %v", err)
	}

	tag, err := domain.NewTag(row.Users.Tag)
	if err != nil {
		return nil, nil, fmt.Errorf("map tag: %v", err)
	}

	displayName, err := domain.NewDisplayName(row.Users.DisplayName)
	if err != nil {
		return nil, nil, fmt.Errorf("map display name: %v", err)
	}

	role, err := domain.NewRole(row.Users.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("map role: %v", err)
	}

	authProvider, err := domain.NewProvider(row.Credentials.Provider)
	if err != nil {
		return nil, nil, fmt.Errorf("map provider: %v", err)
	}

	user := domain.RestoreUser(
		userID,
		email,
		tag,
		displayName,
		role,
		row.Users.CreatedAt,
		row.Users.UpdatedAt,
		fromPgTimestampzPtr(row.Users.DeletedAt),
	)

	credential := domain.RestoreCredential(
		row.Credentials.ID,
		credentialUserID,
		authProvider,
		fromPgTextPtr(row.Credentials.PasswordHash),
		fromPgTextPtr(row.Credentials.ProviderKey),
	)

	return user, credential, nil
}
