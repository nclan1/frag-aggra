package database

import (
	"context"
	"fmt"
	"frag-aggra/internal/models"
	"log"

	"github.com/jackc/pgx/v5"
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

// Close closes the database connection pool.
func (r *Repository) Close() {
	if r != nil && r.dbpool != nil {
		r.dbpool.Close()
	}
}

// querying example
func (r *Repository) QueryRows(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if r.dbpool == nil {
		return nil, fmt.Errorf("database pool is not initialized")
	}
	rows, err := r.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return rows, nil
}

func (r *Repository) InsertItem(ctx context.Context, post models.Post, listing models.FragranceListing) error {
	if r.dbpool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	tx, err := r.dbpool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	//if not committed, rollback
	defer tx.Rollback(ctx)
	// get the unique postID to act as foreign key
	var postID int64

	postInsertQuery := `
		INSERT INTO posts (reddit_id, url, seller_username)
		VALUES ($1, $2, $3)
		ON CONFLICT (reddit_id) DO UPDATE SET
			url = EXCLUDED.url,
			seller_username = EXCLUDED.seller_username
		RETURNING id
	`
	err = tx.QueryRow(ctx, postInsertQuery, post.PostID, post.URL, post.SellerUsername).Scan(&postID)
	if err != nil {
		return fmt.Errorf("failed to insert post: %w", err)
	}

	rows := [][]any{}
	for _, perfume := range listing.Perfumes {
		//iterate through the sizes
		for i, size := range perfume.Sizes {
			if i >= len(perfume.Prices) {
				// Log the inconsistency and break the inner loop for this perfume.
				// This prevents the panic and safely skips the corrupted data.
				log.Printf("warning: mismatched sizes and prices for perfume '%s'. Sizes: %d, Prices: %d. Skipping remaining items.", perfume.Name, len(perfume.Sizes), len(perfume.Prices))
				break
			}
			price := perfume.Prices[i]
			rows = append(rows, []any{postID, perfume.Name, size, price})
		}
	}

	if len(rows) == 0 {
		log.Printf("No valid listing found to insert %s", post.URL)
		return tx.Commit(ctx)
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"listings"},
		[]string{"post_id", "name", "size", "price"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("failed to copy from rows: %w", err)
	}

	return tx.Commit(ctx)
}

// Ping checks if the database connection is alive.
func (r *Repository) Ping(ctx context.Context) error {
	if r.dbpool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return r.dbpool.Ping(ctx)
}
