# FragranceSwap Aggregator (frag-aggra)

a distributed system written in Go that automatically scrapes, parses, and aggregates fragrance sale listings from the `r/fragranceswap` subreddit into a structured, queryable database.

## core problem

i want to find perfumes to buy aftermarket or get decants. Best source is `r/fragranceswap`. Manually going through each listing and searching for the best price size ratio is tedious, hard to track what's for sale and how much it cost. Parsing method for regex is too brittle for the wide variety of post formats. 

this project aims to solve it by building a platform to store a live running database that uses an LLM (via whatever is cheapest llm provider lol, rn using OpenAI) to extract data converting it into a structures JSON. 


## in progress features...

- **automated scraping (in prog):** go service continuously monitors `r/fragranceswap` for new sale posts.
- **intelligent parsing (need some tweaking in prompt):** use an LLM API with a strict JSON schema to extract perfume names, sizes, and prices from raw post text.
- **data persistence:** stores structured data in a PostgreSQL database.
- **distributed architecture (in prog):** decoupled system where scrapers (producers) and parsers (consumers) communicate via a message queue.

## system architecture

producer consumer 

`[scraper service] -> [RabbitMQ message queue] -> [worker service(s)] -> [postgresql database]`

1.  **scraper:** polls the reddit api, finds new sale posts, and publishes a `ParsingJob` to the rabbitmq queue. (again, in progress, no rabbitmq yet)
2.  **worker:** consumes jobs from the queue, sends the post content to the openai api for parsing, and saves the structured result to the postgresql database.

## Technology Stack

- **language:** go
- **data extraction:** gpt-4o
- **database:** psql (with `pgx` driver)
- **message qeue:** rabbitmq (planned)
- **containerization:** docker & docker compose
- **reddit api wrapper:** `go-reddit/v2`

## Getting Started

### Prerequisites

- Go (version 1.23 or later)
- Docker and Docker Compose
- An OpenAI API key
- Reddit API credentials (Client ID, Secret, Username, Password)

### setup

(not yet)

## Testing

The project includes comprehensive unit tests for all packages.

### Run all tests
```bash
go test ./...
```

### Run tests with coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run tests for a specific package
```bash
go test ./internal/parser/... -v
go test ./internal/scraper/... -v
go test ./internal/database/... -v
go test ./internal/models/... -v
```

### Current Test Coverage
- **Parser**: 75.0% - Tests for OpenAI API integration, JSON parsing, and error handling
- **Scraper**: 31.2% - Tests for WTS filtering (100% on containsWTS function)
- **Database**: 27.9% - Tests for error handling and nil pool checks
- **Models**: 100% - Tests for JSON serialization/deserialization

## Project Structure

-   `cmd/`: contains the entry points for the different services (worker, scraper, api).
-   `internal/`: contains all the core application logic, which is not meant to be imported by other projects.
    -   `database/`: handles all communication with the postgresql database.
    -   `parser/`: manages the interaction with the openai api.
    -   `scraper/`: contains the logic for fetching data from reddit.
-   `migrations/`: Holds the sql files for database schema migrations.
-   `docker-compose.yml`: defines the development environment services (postgresql, rabbitmq).
