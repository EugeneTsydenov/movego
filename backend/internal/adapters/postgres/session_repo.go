package postgres

import (
	"context"
	"errors"
	"movego/internal/adapters/postgres/sqlc"
	"movego/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *SessionRepo) FindValid(ctx context.Context, sessionID uuid.UUID) (*domain.Session, *domain.User, error) {
	row, err := r.querier.FindValid(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, err
	}

	return toDomainFindValidRow(row)
}

func (r *SessionRepo) ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	rows, err := r.querier.ListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toDomainSessionList(rows), nil
}

func (r *SessionRepo) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return r.querier.Delete(ctx, sessionID)
}
