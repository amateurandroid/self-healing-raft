# ---- Stage 1: Build the Go binary ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum from the project root (one level up from deploy/)
COPY ../go.mod ../go.sum ./
RUN go mod download

# Copy the rest of the project (again, from project root)
COPY .. .

# Build the server binary
RUN go build -o server ./cmd/server

# ---- Stage 2: Small runtime image ----
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/server /usr/local/bin/server

# Default: will be overridden by docker-compose command args
CMD ["server", "1", "8080"]