package main

import (
	"context"
	"fmt"
	"frag-aggra/internal/database"
	"frag-aggra/internal/parser"
	"frag-aggra/internal/scraper"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Create database connection
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")

	limit := os.Getenv("REDDIT_FETCH_LIMIT")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		log.Printf("invalid REDDIT_FETCH_LIMIT %q, defaulting to 5: %v", limit, err)
		limitInt = 5
	}
	if limitInt <= 0 || limitInt > 100 {
		log.Printf("REDDIT_FETCH_LIMIT %d out of range [1, 100], defaulting to 5", limitInt)
		limitInt = 5
	}

	repo, err := database.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to create repository: %v", err)
	}

	log.Println("Repository created successfully")

	//ping check
	if err := repo.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Database connection verified")
	defer repo.Close()

	scraper, err := scraper.New()
	if err != nil {
		log.Fatalf("failed to initialize reddit scraper: %v", err)
	}

	job_postings, err := scraper.FetchPost("fragranceswap", *repo, limitInt)
	if err != nil {
		log.Fatalf("failed to fetch posts: %v", err)
	}

	p, err := parser.New()
	if err != nil {
		log.Fatalf("failed to create parser: %v", err)
	}

	for _, post := range job_postings {
		raw_input := post.Title + "\n" + post.Body
		log.Printf("Post Title: %s\n", post.Title)
		log.Printf("Post URL: %s\n", post.URL)
		fmt.Println("Parsing Reddit post content...")

		//todo: before parsing, check if post already exists in db by postID
		//...if exists, skip
		//...if not, parse and insert
		parsed_listing, err := p.ParsePostContent(context.Background(), raw_input)
		if err != nil {
			log.Printf("failed to parse post content: %v", err)
			continue
		}
		repo.InsertItem(context.Background(), post, *parsed_listing)
	}

	// query listing
	rows, err := repo.QueryRows(ctx, "SELECT * from listings")
	if err != nil {
		log.Fatalf("failed to query listings: %v", err)
	}
	defer rows.Close()

	log.Println("Listings:")
	for rows.Next() {
		var id int
		var post_id string
		var name string
		var size string
		var price string
		var created_at time.Time
		if err := rows.Scan(&id, &post_id, &name, &size, &price, &created_at); err != nil {
			log.Fatalf("failed to scan row: %v", err)
		}
		log.Printf("ID: %d, Post ID: %s, Name: %s, Size: %s, Price: %s, Created At: %s\n", id, post_id, name, size, price, created_at.Format(time.RFC3339))
	}

}
