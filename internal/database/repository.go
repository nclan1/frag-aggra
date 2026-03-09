package database

import (
	"context"
	"fmt"
	"frag-aggra/internal/models"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Repository struct {
	dbpool *pgxpool.Pool
}

type ListingFilter struct {
	MinPrice float64
	MaxPrice float64
	Search   string // For name search
	SortBy   string // "price_asc", "price_desc", "newest"
	Limit    int
	Offset   int
}

// ListingResponse matches the JSON structure we want for the frontend
type ListingResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Size      string    `json:"size"`
	Price     string    `json:"price"`
	Seller    string    `json:"seller"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
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

func (r *Repository) GetListings(ctx context.Context, filter ListingFilter) ([]ListingResponse, error) {
	if r.dbpool == nil {
		return nil, fmt.Errorf("database pool is not initialized")
	}

	// 1. Start building the base query and argument list
	baseQuery := `
		SELECT
			l.id, l.name, l.size, l.price, p.seller_username, p.url, l.created_at
		FROM listings l
		JOIN posts p ON l.post_id = p.id
		WHERE 1=1
	`

	args := []any{}
	argId := 1
	var conditions []string

	// 2. Add Search Filter (case-insensitive)
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("l.name ILIKE $%d", argId))
		args = append(args, "%"+filter.Search+"%")
		argId++
	}

	// 3. Add Min/Max Price Filters
	// Note: Because price is VARCHAR (e.g. "$120"), we must strip the '$' and cast to NUMERIC.
	// This also assumes all your prices start with '$'. If some are "See Spreadsheet",
	// we need to ignore those when doing numeric comparisons.
	if filter.MinPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("l.price LIKE '$$%%' AND CAST(REPLACE(l.price, '$', '') AS NUMERIC) >= $%d", argId))
		args = append(args, filter.MinPrice)
		argId++
	}

	if filter.MaxPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("l.price LIKE '$$%%' AND CAST(REPLACE(l.price, '$', '') AS NUMERIC) <= $%d", argId))
		args = append(args, filter.MaxPrice)
		argId++
	}

	// 4. Combine conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// 5. Add Sorting
	switch filter.SortBy {
	case "price_asc":
		// Sort text prices ("See Spreadsheet") last, and valid prices numerically ascending
		baseQuery += " ORDER BY (l.price NOT LIKE '$$%'), CAST(NULLIF(REPLACE(l.price, '$', ''), 'See Spreadsheet') AS NUMERIC) ASC"
	case "price_desc":
		baseQuery += " ORDER BY (l.price NOT LIKE '$$%'), CAST(NULLIF(REPLACE(l.price, '$', ''), 'See Spreadsheet') AS NUMERIC) DESC"
	default: // "newest"
		baseQuery += " ORDER BY l.created_at DESC"
	}

	// 6. Add Pagination (Limit & Offset)
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argId)
		args = append(args, filter.Limit)
		argId++
	}

	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argId)
		args = append(args, filter.Offset)
		argId++
	}

	// 7. Execute the query
	rows, err := r.dbpool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query listings: %w", err)
	}
	defer rows.Close()

	// 8. Scan the results into the slice
	var listings []ListingResponse
	for rows.Next() {
		var l ListingResponse
		err := rows.Scan(
			&l.ID,
			&l.Name,
			&l.Size,
			&l.Price,
			&l.Seller,
			&l.URL,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan listing row: %w", err)
		}
		listings = append(listings, l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Return empty array instead of nil if no listings found (better for JSON encoding)
	if listings == nil {
		listings = []ListingResponse{}
	}

	return listings, nil
}

// Close closes the database connection pool.
func (r *Repository) Close() {
	if r != nil && r.dbpool != nil {
		r.dbpool.Close()
	}
}

func (r *Repository) PostExists(ctx context.Context, redditID string) (bool, error) {
	if r.dbpool == nil {
		return false, fmt.Errorf("database pool is not initialized")
	}
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE reddit_id=$1)`
	err := r.dbpool.QueryRow(ctx, query, redditID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}
	return exists, nil
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
