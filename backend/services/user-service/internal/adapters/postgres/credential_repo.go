package postgres

import (
	"context"
	"errors"
	"user/internal/adapters/postgres/sqlc"
	"user/internal/domain"

	"github.com/jackc/pgx/v5"
)

var _ domain.CredentialRepo = (*CredentialRepo)(nil)

type CredentialRepo struct {
	querier sqlc.Querier
}

func NewCredentialRepo(querier sqlc.Querier) *CredentialRepo {
	return &CredentialRepo{
		querier: querier,
	}
}

func (r *CredentialRepo) Save(ctx context.Context, cred *domain.Credential) error {
	err := r.querier.SaveCredential(ctx, toSaveCredentialsParams(cred))
	return mapCredentialError(err)
}

func (r *CredentialRepo) FindForAuth(
	ctx context.Context,
	email domain.Email,
	provider domain.Provider,
) (*domain.User, *domain.Credential, error) {
	row, err := r.querier.FindForAuth(ctx, toFindForAuthParams(email, provider))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}

		return nil, nil, err
	}

	user, credential, err := toDomainFindForAuthRow(row)
	if err != nil {
		return nil, nil, err
	}
	return user, credential, nil
}
