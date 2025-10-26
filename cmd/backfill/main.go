package main

import (
	"context"
	"frag-aggra/internal/models"
	"frag-aggra/internal/pubsub"
	"frag-aggra/internal/scraper"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {

	log.Println("Starting one-time backfill job...")

	_ = godotenv.Load()
	ctx := context.Background()
	rmqUrl := os.Getenv("RABBITMQ_URL")
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
	// Possible TODO: defer close the scraper, look at the documentation.

	rmq, err := pubsub.New(rmqUrl)
	if err != nil {
		log.Fatalf("Failed to innit RabbitMQ Client: %v", err)
	}
	defer rmq.Close()

	// jobPostings, err := scraper.FetchPost("fragranceswap", *repo, limitInt)
	// if err != nil {
	// 	log.Fatalf("Failed to fetch posts: %v", err)
	// }

	// grab a cut off date
	cutoffDate := time.Now().Add(-14 * 24 * time.Hour) // 2 weeks ago
	maxPostLimit := 2000
	totalPublished := 0
	afterToken := ""

	log.Printf("Starting backfill. Cutoff date: %s, Max posts: %d", cutoffDate.Format(time.RFC3339), maxPostLimit)

	for totalPublished < maxPostLimit {
		log.Printf("Fetching page of posts (after: %s)...", afterToken)

		posts, err := scraper.FetchPaginatedPosts(ctx, "fragranceswap", limitInt, afterToken)
		if err != nil {
			log.Printf("Error fetching historical posts, stopping: %v", err)
			break
		}
		if len(posts) == 0 {
			log.Println("No more posts found, exiting...")
			break
		}

		hitCutoffDate := false

		for _, post := range posts {

			//hit cut off date eyt
			if post.Created.Time.Before(cutoffDate) {
				log.Printf("Hit time cut-off, stopping backfill")
				hitCutoffDate = true
				break
			}

			//check if past the upperbound fallback
			if totalPublished >= maxPostLimit {
				log.Printf("Parsed more than %d posts, stopping", maxPostLimit)
				hitCutoffDate = true
				break
			}

			//check if post has the wts tag to filter it out
			if !scraper.ContainsWTS(post.Title) && !scraper.ContainsWTS(post.Body) {
				log.Printf("Skipping post %s without [WTS] in title or body", post.ID)
				continue
			}

			job_post := models.Post{
				PostID:         post.ID,
				URL:            post.URL,
				Title:          post.Title,
				Body:           post.Body,
				SellerUsername: post.Author,
			}

			err := rmq.Publish2JSON("exchange_here", "key_here", job_post, ctx)
			if err != nil {
				log.Printf("Error publishing post to RabbitMQ client with post ID %s: %v", post.ID, err)
			} else {
				totalPublished++
			}
		}

		if hitCutoffDate {
			break
		}
		afterToken = posts[len(posts)-1].ID
		log.Printf("Published %d jobs so far. Sleeping for 2s...", totalPublished)
		time.Sleep(2 * time.Second) // Being nice to Reddit's API
	}

	log.Printf("Backfill complete. Published %d total jobs", totalPublished)

}
