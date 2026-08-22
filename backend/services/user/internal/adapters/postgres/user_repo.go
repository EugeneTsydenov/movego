package postgres

import (
	"context"
	"user/internal/adapters/postgres/sqlc"
	"user/internal/domain"
)

var _ domain.UserRepo = (*UserRepo)(nil)

type UserRepo struct {
	querier sqlc.Querier
}

func NewUserRepo(querier sqlc.Querier) *UserRepo {
	return &UserRepo{
		querier: querier,
	}
}

func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	err := r.querier.SaveUser(ctx, toSaveUsersParams(user))
	return mapUserError(err)
}
