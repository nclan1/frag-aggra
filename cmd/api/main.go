package main

import (
	"context"
	"encoding/json"
	"fmt"
	"frag-aggra/internal/database"
	"log"
	"net/http"
	"os"

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
