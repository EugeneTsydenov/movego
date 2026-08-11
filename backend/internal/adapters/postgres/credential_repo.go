package postgres

import (
	"context"
	"movego/internal/adapters/postgres/sqlc"
	"movego/internal/domain"
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
