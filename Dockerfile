# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/api/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

# Install stockfish for the engine service
RUN apk add --no-cache stockfish

# Copy the binary from the builder stage
COPY --from=builder /app/main .
COPY .env .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
