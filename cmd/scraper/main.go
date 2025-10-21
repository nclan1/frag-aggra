package main

import (
	"context"
	"frag-aggra/internal/scraper"
	"log"

	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	ctx := context.Background()

	scraper, err := scraper.New()

	if err != nil {
		log.Fatalf("Failed to init reddit scraper: %v", err)
	}

	job_postings, err := scraper.

}
