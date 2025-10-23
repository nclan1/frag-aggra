package main

import (
	"context"
	"frag-aggra/internal/database"
	"frag-aggra/internal/scraper"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {

	log.Println("Starting one-time backfill job...")

	_ = godotenv.Load()
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	limit := os.Getenv("REDDIT_FETCH_LIMIT")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		log.Printf("invalid REDDIT_FETCH_LIMIT configuration %q, defaulting to 5: %v", limit, err)
		limitInt = 5
	}
	if limitInt <= 0 || limitInt > 100 {
		log.Printf("REDDIT_FETCH_LIMIT %d out of range [1, 100], defaulting to 5", limitInt)
		limitInt = 5
	}

	// connect to db
	repo, err := database.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to repository: %v", err)
	}

	// init a scraper
	scraper, err := scraper.New()
	if err != nil {
		log.Fatalf("Failed to init reddit scraper: %v", err)
	}

	// Parse however much and input it into job_postings
	job_postings, err := scraper.FetchPost("fragranceswap", *repo, limitInt)
	if err != nil {
		log.Fatalf("Failed to fetch posts: %v", err)
	}

}
