package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Repo() Repository {
	return &postgresRepository{db: s.pool}
}

func (s *Store) WithTx(ctx context.Context, fn func(repo Repository) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := fn(&postgresRepository{db: tx}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
