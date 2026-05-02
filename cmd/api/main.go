package main

import (
	"context"
	"encoding/json"
	"fmt"
	"frag-aggra/internal/database"
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
