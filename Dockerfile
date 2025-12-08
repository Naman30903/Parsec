# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make wget

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o processor ./cmd/processor

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and wget for healthcheck
RUN apk --no-cache add ca-certificates wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/processor .

# Expose port
EXPOSE 8080

# Run the processor
CMD ["./processor"]
