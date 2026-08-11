package postgres

import (
	"context"
	"movego/internal/adapters/postgres/sqlc"
	"movego/internal/domain"
)

var _ domain.SessionRepo = (*SessionRepo)(nil)

type SessionRepo struct {
	querier sqlc.Querier
}

func NewSessionRepo(querier sqlc.Querier) *SessionRepo {
	return &SessionRepo{
		querier: querier,
	}
}

func (r *SessionRepo) Save(ctx context.Context, session *domain.Session) error {
	err := r.querier.SaveSession(ctx, toSaveSessionsParams(session))
	return mapSessionError(err)
}
