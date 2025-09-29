package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Repository struct {
	dbpool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*Repository, error) {
	if connString == "" {
		return nil, fmt.Errorf("connection string is empty")
	}

	//pgxpool.New is sql.Open + db.Ping()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &Repository{dbpool: pool}, nil
}

func (r *Repository) Close() {
	if r != nil && r.dbpool != nil {
		r.dbpool.Close()
	}
}

func (r *Repository) Ping(ctx context.Context) error {
	if r.dbpool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return r.dbpool.Ping(ctx)
}
