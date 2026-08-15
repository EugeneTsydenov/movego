package postgres

import (
	"errors"
	"movego/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

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
