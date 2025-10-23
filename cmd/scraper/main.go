package main

import (
	"frag-aggra/internal/scraper"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {

	// all it does here is poll reddit, and publish each message to rabbitmq queue for worker to consume

	// TODO: needs to connect to RabbitMQ, set up here

	// TODO: set up a ticking timer for x amount of seconds that loops.

	pollInterval := 5 * time.Minute
	ticker := time.NewTicker(pollInterval)

	_ = godotenv.Load()
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

	// init a scraper
	scraper, err := scraper.New()
	if err != nil {
		log.Fatalf("Failed to init reddit scraper: %v", err)
	}

	log.Println("Scraper service started. Polling every 5 minutes")
	for {
		<-ticker.C

		log.Println("Polling latest reddit posts")
		// Parse however much and input it into job_postings
		job_postings, err := scraper.FetchPost("fragranceswap", limitInt)
		if err != nil {
			log.Fatalf("Failed to fetch posts: %v", err)
		}

	}
}
