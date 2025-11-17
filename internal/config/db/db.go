package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(ctx context.Context, dns string) (*DB, error) {
	dbPool, err := pgxpool.New(ctx, dns)
	if err != nil {
		return nil, err
	}

	if err := dbPool.Ping(ctx); err != nil {
		return nil, err
	}

	return &DB{Pool: dbPool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
