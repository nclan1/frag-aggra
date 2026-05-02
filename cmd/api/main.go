package main

import (
	"context"
	"encoding/json"
	"fmt"
	"frag-aggra/internal/database"
	"frag-aggra/internal/pubsub"
	"frag-aggra/internal/routing"
	"frag-aggra/internal/scraper"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	// Recoverer prevents the server from crashing if a request panics
	r.Use(middleware.Recoverer)

	//set up database repository
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

	// Optional: Reddit scraper + RabbitMQ for the search endpoint
	var sc *scraper.RedditScraper
	var rmq *pubsub.RabbitMQClient

	if os.Getenv("REDDIT_CLIENT_ID") != "" {
		if s, err := scraper.New(); err != nil {
			log.Printf("warning: failed to init reddit scraper: %v", err)
		} else {
			sc = s
		}
	}

	if rmqURL := os.Getenv("RABBITMQ_URL"); rmqURL != "" && sc != nil {
		if r, err := pubsub.New(rmqURL); err != nil {
			log.Printf("warning: failed to init RabbitMQ client: %v", err)
		} else {
			rmq = r
			defer rmq.Close()
			if err := rmq.Channel.ExchangeDeclare(routing.ExchangePostDirect, "direct", true, false, false, false, nil); err != nil {
				log.Printf("warning: failed to declare exchange: %v", err)
				rmq = nil
			} else if q, err := rmq.Channel.QueueDeclare(routing.PostQueue, true, false, false, false, nil); err != nil {
				log.Printf("warning: failed to declare queue: %v", err)
				rmq = nil
			} else if err := rmq.Channel.QueueBind(q.Name, routing.PostKey, routing.ExchangePostDirect, false, nil); err != nil {
				log.Printf("warning: failed to bind queue: %v", err)
				rmq = nil
			} else {
				log.Println("Search feature ready")
			}
		}
	}

	r.Route("/api/v1", func(r chi.Router) {

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		r.Get("/listings", func(w http.ResponseWriter, r *http.Request) {
			// 1. Parse Query Parameters
			query := r.URL.Query()

			limit, _ := strconv.Atoi(query.Get("limit"))
			if limit <= 0 {
				limit = 20
			} // default limit

			page, _ := strconv.Atoi(query.Get("page"))
			if page <= 0 {
				page = 1
			} // default page

			minPrice, _ := strconv.ParseFloat(query.Get("min_price"), 64)
			maxPrice, _ := strconv.ParseFloat(query.Get("max_price"), 64)

			// 2. Construct Filter
			filter := database.ListingFilter{
				Search:   query.Get("q"),
				MinPrice: minPrice,
				MaxPrice: maxPrice,
				SortBy:   query.Get("sort"),
				Limit:    limit,
				Offset:   (page - 1) * limit,
			}

			// 3. Fetch from DB
			listings, err := repo.GetListings(r.Context(), filter)
			if err != nil {
				http.Error(w, "Failed to fetch listings", http.StatusInternalServerError)
				return
			}

			total, err := repo.CountListings(r.Context(), filter)
			if err != nil {
				log.Printf("failed to count listings: %v", err)
			}

			// 4. Return standardized JSON response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data": listings,
				"meta": map[string]any{
					"page":        page,
					"limit":       limit,
					"total_count": total,
				},
			})
		})

		r.Post("/search", func(w http.ResponseWriter, r *http.Request) {
			if sc == nil || rmq == nil {
				http.Error(w, "search not available: missing reddit or rabbitmq config", http.StatusServiceUnavailable)
				return
			}

			var req struct {
				Keyword string `json:"keyword"`
				Limit   int    `json:"limit"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Keyword == "" {
				http.Error(w, "invalid request: keyword required", http.StatusBadRequest)
				return
			}
			if req.Limit <= 0 || req.Limit > 100 {
				req.Limit = 25
			}

			posts, err := sc.SearchPosts(r.Context(), "fragranceswap", req.Keyword, req.Limit)
			if err != nil {
				log.Printf("failed to search reddit: %v", err)
				http.Error(w, "failed to search reddit", http.StatusInternalServerError)
				return
			}

			queued := 0
			for _, post := range posts {
				if err := rmq.Publish2JSON(routing.ExchangePostDirect, routing.PostKey, post, r.Context()); err != nil {
					log.Printf("failed to publish post %s: %v", post.PostID, err)
				} else {
					queued++
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"queued":  queued,
				"found":   len(posts),
				"keyword": req.Keyword,
			})
		})
	})
	// 4. Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the Frag Aggra API!"))
	})

	r.Get("/listings", func(w http.ResponseWriter, r *http.Request) {

		//query the database
	})

	r.Get("/items", func(w http.ResponseWriter, r *http.Request) {
		// In a real app, you would query your Postgres DB here
		items := []Item{
			{ID: "1", Name: "Fragrance A"},
			{ID: "2", Name: "Aggregator B"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	})

	// 5. Start the Server
	// IMPORTANT: This must match the EXPOSE 8080 in your Dockerfile
	fmt.Println("Server running on port 8080...")
	http.ListenAndServe(":8080", r)
}
