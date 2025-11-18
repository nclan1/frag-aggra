package main

import (
	"context"
	"encoding/json"
	"frag-aggra/internal/database"
	"frag-aggra/internal/models"
	"frag-aggra/internal/parser"
	"frag-aggra/internal/pubsub"
	"frag-aggra/internal/routing"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Create database connection
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")

	rmqUrl := os.Getenv("RABBITMQ_URL")
	if rmqUrl == "" {
		log.Fatal("RABBITMQ_URL not set")
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

	log.Printf("Initializing the parser...")
	p, err := parser.New()
	if err != nil {
		log.Fatalf("failed to create parser: %v", err)
	}
	log.Println("Parser created successfully")

	// connect to the rabbitmq
	log.Print("Connecting to RabbitMQ...")
	rmq, err := pubsub.New(rmqUrl)
	if err != nil {
		log.Fatalf("Failed to innit RabbitMQ Client: %v", err)
	}
	log.Print("RabbitMQ connection established")
	defer rmq.Close()

	queue := routing.PostQueue
	//declare the queue
	q, err := rmq.Channel.QueueDeclare(
		queue, //queue name
		true,  //durable
		false, //delete when unused
		false, //exclusive
		false, //nowait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	log.Printf("Grabbing 1 post from queue in rabbitmq")

	msgs, err := rmq.ConsumeFromClient(q.Name)
	if err != nil {
		log.Fatalf("Error consuming and getting channel")
	}

	// add signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for msg := range msgs {
			// msg := <-msgs
			jsonStr := string(msg.Body)
			var post models.Post
			if err := json.Unmarshal([]byte(jsonStr), &post); err != nil {
				log.Printf("bad json: %v", err)
				msg.Nack(false, false) //drop, dont requeu garbage
				continue
			}
			// for _, post := range job_postings {
			raw_input := post.Title + "\n" + post.Body
			log.Printf("Post Title: %s\n", post.Title)
			log.Printf("Post URL: %s\n", post.URL)
			log.Printf("Post id: %s\n", post.PostID)
			log.Printf("Parsing Reddit post content...")

			exists, err := repo.PostExists(ctx, post.PostID)
			if err != nil {
				log.Printf("Error checking existence")
			}
			if !exists {
				parsed_listing, err := p.ParsePostContent(context.Background(), raw_input)
				if err != nil {
					log.Printf("failed to parse post content: %v", err)
					msg.Nack(false, true) // re-queue
					continue
				}
				if parsed_listing == nil {
					log.Printf("parser returned nil for post %s", post.PostID)
					msg.Nack(false, true)
					continue
				}
				repo.InsertItem(ctx, post, *parsed_listing)
				log.Printf("Finished parsing post %s", post.PostID)
			} else {
				log.Printf("Already seen, skipping")
			}
			msg.Ack(false)
		}
	}()
	<-sigChan
	log.Println("Shutting down gracefully...")

}
