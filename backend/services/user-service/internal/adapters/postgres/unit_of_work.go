package postgres

import (
	"context"
	"fmt"
	"user/internal/adapters/postgres/sqlc"
	"user/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(repos domain.Repos) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)
	repos := &repositories{
		users:       NewUserRepo(q),
		credentials: NewCredentialRepo(q),
		sessions:    NewSessionRepo(q),
	}

	if err := fn(repos); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type repositories struct {
	users       *UserRepo
	credentials *CredentialRepo
	sessions    *SessionRepo
}

func (r *repositories) Users() domain.UserRepo {
	return r.users
}

func (r *repositories) Credentials() domain.CredentialRepo {
	return r.credentials
}

func (r *repositories) Sessions() domain.SessionRepo {
	return r.sessions
}
