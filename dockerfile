# 1. Update the base image to match your go.mod version (1.23)
FROM golang:1.23-alpine as builder

# 2. Use a clean, standard directory inside the container
WORKDIR /app

# 3. Copy dependencies first (caching layer)
COPY go.mod go.sum ./
RUN go mod download

# 4. Copy the source code
COPY . .

# 5. Build the binary
# We name the binary "frag-aggra" to match your module name
RUN CGO_ENABLED=0 GOOS=linux go build -o /frag-aggra ./cmd/api

# --- Final Stage ---
FROM alpine:latest

# Install certificates for your OpenAI and Reddit API calls
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder
COPY --from=builder /frag-aggra /usr/bin/frag-aggra

EXPOSE 8080

# Run it
ENTRYPOINT ["/usr/bin/frag-aggra"]
