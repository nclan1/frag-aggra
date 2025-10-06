package main

import (
	"context"
	"fmt"
	"frag-aggra/internal/database"
	"frag-aggra/internal/parser"
	"frag-aggra/internal/scraper"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// 	p, err := parser.New()
	// 	if err != nil {
	// 		log.Fatalf("failed to create parser: %v", err)
	// 	}

	// 	fmt.Println("Parsing Reddit post content...")
	// 	listing, err := p.ParsePostContent(context.Background(), redditPost)
	// 	if err != nil {
	// 		log.Fatalf("failed to parse post content: %v", err)
	// 	}
	// 	fmt.Println("\nPARSED OUTPUT:")
	// 	for _, perfume := range listing.Perfumes {
	// 		fmt.Printf("Name: %s\n", perfume.Name)
	// 		for i, size := range perfume.Sizes {
	// 			price := perfume.Prices[i]
	// 			fmt.Printf("  Size: %s - Price: %s\n", size, price)

	//		}
	//	}

	// Create database connection
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
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

	job_postings, err := scraper.FetchPost("fragranceswap")
	if err != nil {
		log.Fatalf("failed to fetch posts: %v", err)
	}

	raw_input := job_postings[0].Body

	p, err := parser.New()
	if err != nil {
		log.Fatalf("failed to create parser: %v", err)
	}

	// Parse the post content
	parsed_listing, err := p.ParsePostContent(context.Background(), raw_input)
	if err != nil {
		log.Fatalf("failed to parse post content: %v", err)
	}
	fmt.Println("\nPARSED OUTPUT:")
	for _, perfume := range parsed_listing.Perfumes {
		fmt.Printf("Name: %s\n", perfume.Name)
		for i, size := range perfume.Sizes {
			price := perfume.Prices[i]
			fmt.Printf("  Size: %s - Price: %s\n", size, price)
		}
	}

	// fmt.Println("Parsing Reddit post content...")
	// listing, err := p.ParsePostContent(context.Background(), redditPost)
	// if err != nil {
	// 	log.Fatalf("failed to parse post content: %v", err)
	// }

	//query listing
	// rows, err := repo.QueryRows(ctx, "SELECT * from listings")
	// if err != nil {
	// 	log.Fatalf("failed to query listings: %v", err)
	// }
	// defer rows.Close()

	// log.Println("Listings:")
	// for rows.Next() {
	// 	var id int
	// 	var post_id string
	// 	var name string
	// 	var size string
	// 	var price string
	// 	var created_at time.Time
	// 	if err := rows.Scan(&id, &post_id, &name, &size, &price, &created_at); err != nil {
	// 		log.Fatalf("failed to scan row: %v", err)
	// 	}
	// 	log.Printf("ID: %d, Post ID: %s, Name: %s, Size: %s, Price: %s, Created At: %s\n", id, post_id, name, size, price, created_at.Format(time.RFC3339))
	// }

}

// func buildConnString() string {

// }
