package main

import (
	"context"
	"fmt"
	"frag-aggra/internal/database"
	"frag-aggra/internal/pubsub"
	"frag-aggra/internal/routing"
	"log"
	"os"

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
	// p, err := parser.New()
	// if err != nil {
	// 	log.Fatalf("failed to create parser: %v", err)
	// }
	// log.Println("Parser created successfully")

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
		log.Fatalf("Error consuming and getting channel", err)
	}

	d := <-msgs
	fmt.Printf("body: %s\n", d.Body)

	// for _, post := range job_postings {
	// 	raw_input := post.Title + "\n" + post.Body
	// 	log.Printf("Post Title: %s\n", post.Title)
	// 	log.Printf("Post URL: %s\n", post.URL)
	// 	fmt.Println("Parsing Reddit post content...")

	// 	//todo: before parsing, check if post already exists in db by postID
	// 	//...if exists, skip
	// 	//...if not, parse and insert
	// 	parsed_listing, err := p.ParsePostContent(context.Background(), raw_input)
	// 	if err != nil {
	// 		log.Printf("failed to parse post content: %v", err)
	// 		continue
	// 	}
	// 	repo.InsertItem(context.Background(), post, *parsed_listing)
	// }

	// query listing
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
