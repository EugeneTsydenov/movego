package postgres

import (
	"movego/internal/adapters/postgres/sqlc"
	"movego/internal/domain"
	"time"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jackc/pgx/v5/pgconn"
)

// pg <-> domain
func toPgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

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

func toPgTimestampz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
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

func toSaveCredentialsParams(cred *domain.Credential) sqlc.SaveCredentialParams {
	return sqlc.SaveCredentialParams{
		ID:           toPgUUID(cred.ID()),
		UserID:       toPgUUID(cred.UserID()),
		PasswordHash: toPgTextPtr(cred.PasswordHash()),
		Provider:     cred.Provider().String(),
		ProviderKey:  toPgTextPtr(cred.ProviderKey()),
	}
}

func toSaveSessionsParams(session *domain.Session) sqlc.SaveSessionParams {
	return sqlc.SaveSessionParams{
		ID:         toPgUUID(session.ID()),
		UserID:     toPgUUID(session.UserID()),
		SecretHash: session.SecretHash(),
		UserAgent:  session.UserAgent(),
		ClientIp:   session.ClientIP(),
		ExpiresAt:  toPgTimestampz(session.ExpiresAt()),
		CreatedAt:  toPgTimestampz(session.CreatedAt()),
	}
}

func toSaveUsersParams(user *domain.User) sqlc.SaveUserParams {
	return sqlc.SaveUserParams{
		ID:          toPgUUID(user.ID()),
		Email:       user.Email().String(),
		Tag:         user.Tag().String(),
		DisplayName: user.DisplayName().String(),
		Role:        user.Role().String(),
		CreatedAt:   toPgTimestampz(user.CreatedAt()),
		UpdatedAt:   toPgTimestampz(user.UpdatedAt()),
		DeletedAt:   toPgTimestampzPtr(user.DeletedAt()),
	}
}

// handle error
func mapCredentialError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrCredentialNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			switch pgErr.ConstraintName {
			case "credentials_provider_user_id_uq":
				return domain.ErrProviderAlreadyLinked
			case "credentials_provider_key_uq":
				return domain.ErrProviderKeyTaken
			}
		case "23503": // foreign_key_violation
			return domain.ErrUserNotFound
		case "23514": // check_violation
			return domain.ErrInvalidProvider
		}
	}

	return err
}

func mapSessionError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrSessionNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation
			return domain.ErrUserNotFound
		}
	}

	return err
}

func mapUserError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			switch pgErr.ConstraintName {
			case "users_email_key":
				return domain.ErrEmailTaken
			case "users_tag_key":
				return domain.ErrTagTaken
			}
		}
	}

	return err
}
