package main

import (
	"context"
	"frag-aggra/internal/pubsub"
	"frag-aggra/internal/routing"
	"frag-aggra/internal/scraper"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	pollInterval := 5 * time.Minute
	if v := os.Getenv("POLL_INTERVAL_SECONDS"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			pollInterval = time.Duration(secs) * time.Second
		}
	}

	limitInt := 25
	if v := os.Getenv("REDDIT_FETCH_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limitInt = n
		} else {
			log.Printf("invalid REDDIT_FETCH_LIMIT %q, using default %d", v, limitInt)
		}
	}

	sc, err := scraper.New()
	if err != nil {
		log.Fatalf("failed to init reddit scraper: %v", err)
	}

	rmqUrl := os.Getenv("RABBITMQ_URL")
	if rmqUrl == "" {
		log.Fatal("RABBITMQ_URL not set")
	}

	rmq, err := pubsub.New(rmqUrl)
	if err != nil {
		log.Fatalf("failed to init RabbitMQ client: %v", err)
	}
	defer rmq.Close()

	exchange := routing.ExchangePostDirect
	key := routing.PostKey
	queue := routing.PostQueue

	if err := rmq.Channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}
	q, err := rmq.Channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}
	if err := rmq.Channel.QueueBind(q.Name, key, exchange, false, nil); err != nil {
		log.Fatalf("failed to bind queue: %v", err)
	}

	ctx := context.Background()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Printf("scraper started, polling every %s with limit %d", pollInterval, limitInt)

	// poll once immediately on startup, then on each tick
	poll(ctx, sc, rmq, exchange, key, limitInt)
	for range ticker.C {
		poll(ctx, sc, rmq, exchange, key, limitInt)
	}
}

func poll(ctx context.Context, sc *scraper.RedditScraper, rmq *pubsub.RabbitMQClient, exchange, key string, limit int) {
	log.Println("polling r/fragranceswap...")
	posts, err := sc.FetchPost("fragranceswap", limit)
	if err != nil {
		log.Printf("failed to fetch posts: %v", err)
		return
	}

	published := 0
	for _, post := range posts {
		if err := rmq.Publish2JSON(exchange, key, post, ctx); err != nil {
			log.Printf("failed to publish post %s: %v", post.PostID, err)
		} else {
			published++
		}
	}
	log.Printf("published %d/%d posts to queue", published, len(posts))
}
