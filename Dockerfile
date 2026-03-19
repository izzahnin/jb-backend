# Multi-stage build for smaller image size
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./

# Download modules
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o jb-api cmd/api/main.go

# Final stage - runtime image
FROM alpine:latest

WORKDIR /root/

# Install runtime dependencies (ca-certificates for HTTPS)
RUN apk --no-cache add ca-certificates postgresql-client

# Copy binary from builder
COPY --from=builder /app/jb-api .

# Copy docs for Swagger UI
COPY --from=builder /app/docs ./docs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD nc -z localhost 8080 || exit 1

# Run the application
CMD ["./jb-api"]
