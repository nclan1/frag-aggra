package database

import (
	"context"
	"frag-aggra/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_EmptyConnectionString(t *testing.T) {
	ctx := context.Background()
	repo, err := New(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "connection string is empty")
}

func TestNew_InvalidConnectionString(t *testing.T) {
	ctx := context.Background()
	repo, err := New(ctx, "invalid-connection-string")

	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestRepository_Close(t *testing.T) {
	// Test closing nil repository
	var repo *Repository
	repo.Close() // Should not panic
	
	// Test closing repository with nil pool
	repo = &Repository{dbpool: nil}
	repo.Close() // Should not panic
}

func TestRepository_Ping_NilPool(t *testing.T) {
	repo := &Repository{dbpool: nil}
	err := repo.Ping(context.Background())
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database pool is not initialized")
}

func TestRepository_QueryRows_NilPool(t *testing.T) {
	repo := &Repository{dbpool: nil}
	rows, err := repo.QueryRows(context.Background(), "SELECT * FROM listings")
	
	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Contains(t, err.Error(), "database pool is not initialized")
}

func TestRepository_InsertItem_NilPool(t *testing.T) {
	repo := &Repository{dbpool: nil}
	
	post := models.Post{
		PostID:         "test123",
		URL:            "https://reddit.com/test123",
		Title:          "[WTS] Test",
		Body:           "Test body",
		SellerUsername: "testuser",
	}
	
	listing := models.FragranceListing{
		Perfumes: []models.Perfume{
			{
				Name:   "Tom Ford Tobacco Vanille",
				Sizes:  []string{"100ml"},
				Prices: []string{"$150"},
			},
		},
	}
	
	err := repo.InsertItem(context.Background(), post, listing)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database pool is not initialized")
}

// Note: Integration tests with actual database would be ideal but require
// a test database setup. The tests above cover the error paths and nil checks.
// For full coverage of database operations, consider:
// 1. Using docker-compose with a test database in CI
// 2. Using testcontainers for integration tests
// 3. Setting up a separate test database configuration

