package postgres

import (
	"context"
	"errors"
	"shared/coreerrors"
	"user/internal/adapters/postgres/sqlc"
	"user/internal/domain"

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

func (r *SessionRepo) FindValid(ctx context.Context, sessionID domain.SessionID) (*domain.Session, *domain.User, error) {
	row, err := r.querier.FindValid(ctx, sessionID.UUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, coreerrors.ErrNotFound
		}
		return nil, nil, err
	}

	session, user, err := toDomainFindValidRow(row)
	if err != nil {
		return nil, nil, err
	}
	return session, user, nil
}

func (r *SessionRepo) FindByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	session, err := r.querier.FindByID(ctx, id.UUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}

		return nil, err
	}

	s, err := toDomainSession(session)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r *SessionRepo) ListActiveByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error) {
	rows, err := r.querier.ListActiveByUserID(ctx, userID.UUID())
	if err != nil {
		return nil, err
	}

	sessions, err := toDomainSessionList(rows)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepo) Delete(ctx context.Context, sessionID domain.SessionID) error {
	return r.querier.Delete(ctx, sessionID.UUID())
}

func (r *SessionRepo) DeleteAllExcept(ctx context.Context, userID domain.UserID, sessionID domain.SessionID) error {
	return r.querier.DeleteAllExcept(ctx, sqlc.DeleteAllExceptParams{
		ID:     sessionID.UUID(),
		UserID: userID.UUID(),
	})
}
