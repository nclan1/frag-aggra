# 1. Update the base image to match your go.mod version (1.23)
FROM golang:1.23-alpine as builder

ARG CMD=api

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /frag-app ./cmd/${CMD}

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /frag-app /usr/bin/frag-app

EXPOSE 8080

ENTRYPOINT ["/usr/bin/frag-app"]
