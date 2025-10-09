# FragranceSwap Aggregator (frag-aggra)

A distributed system written in Go that automatically scrapes, parses, and aggregates fragrance sale listings from the `r/fragranceswap` subreddit into a structured, queryable database.

## Core Problem

Listings on `r/fragranceswap` are unstructured, user-generated text, making it difficult to track what's for sale, by whom, and for how much. Traditional parsing methods like regex are too brittle to handle the wide variety of post formats.

This project solves the problem by using an LLM (via the OpenAI API) to perform intelligent data extraction, converting messy text into clean, structured JSON that can be stored and analyzed.

## Features

- **Automated Scraping:** A Go service continuously monitors `r/fragranceswap` for new sale posts.
- **Intelligent Parsing:** Leverages the OpenAI API with a strict JSON schema to extract perfume names, sizes, and prices from raw post text.
- **Data Persistence:** Stores structured data in a PostgreSQL database.
- **Distributed Architecture:** A decoupled system where scrapers (producers) and parsers (consumers) communicate via a message queue.

## System Architecture

The system is designed using a producer-consumer pattern to ensure scalability and resilience.

`[Scraper Service] -> [RabbitMQ Message Queue] -> [Worker Service(s)] -> [PostgreSQL Database]`

1.  **Scraper:** Polls the Reddit API, finds new sale posts, and publishes a `ParsingJob` to the RabbitMQ queue.
2.  **Worker:** Consumes jobs from the queue, sends the post content to the OpenAI API for parsing, and saves the structured result to the PostgreSQL database.

## Technology Stack

- **Language:** Go
- **Data Extraction:** OpenAI API (GPT-4o)
- **Database:** PostgreSQL (with `pgx` driver)
- **Message Queue:** RabbitMQ (Planned)
- **Containerization:** Docker & Docker Compose
- **Reddit API Wrapper:** `go-reddit/v2`

## Getting Started

### Prerequisites

- Go (version 1.23 or later)
- Docker and Docker Compose
- An OpenAI API key
- Reddit API credentials (Client ID, Secret, Username, Password)

### Setup

1.  **Clone the repository:**
    ```bash
    git clone <your-repo-url>
    cd frag-aggra
    ```

2.  **Create an environment file:**
    Create a `.env` file in the root of the project by copying the example below. Fill in your own credentials.

    ```ini
    # .env.example
    # OpenAI
    OPENAI_API_KEY="sk-..."

    # Reddit API Credentials
    REDDIT_CLIENT_ID="..."
    REDDIT_CLIENT_SECRET="..."
    REDDIT_USERNAME="..."
    REDDIT_PASSWORD="..."

    # PostgreSQL Connection
    DATABASE_URL="postgresql://postgres:mysecurepassword@localhost:5432/fragrance_database"
    POSTGRES_PASSWORD="mysecurepassword" # Must match the password in docker-compose.yml
    ```

3.  **Start the database:**
    ```bash
    docker-compose up -d
    ```

4.  **Run database migrations:**
    You will need a migration tool like `golang-migrate` to apply the schema.
    ```bash
    migrate -path migrations -database "$DATABASE_URL" up
    ```

### Running the Services

-   **Run the Worker:**
    ```bash
    go run ./cmd/worker/main.go
    ```
-   **Run the Scraper (once created):**
    ```bash
    go run ./cmd/scraper/main.go
    ```

## Project Structure

-   `cmd/`: Contains the entry points for the different services (worker, scraper, api).
-   `internal/`: Contains all the core application logic, which is not meant to be imported by other projects.
    -   `database/`: Handles all communication with the PostgreSQL database.
    -   `parser/`: Manages the interaction with the OpenAI API.
    -   `scraper/`: Contains the logic for fetching data from Reddit.
-   `migrations/`: Holds the SQL files for database schema migrations.
-   `docker-compose.yml`: Defines the development environment services (Postgres, RabbitMQ).
