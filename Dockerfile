# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/api

# Runtime stage
FROM alpine:3.18

# Install ca-certificates for HTTPS, Docker CLI for service discovery, and wget for healthcheck
RUN apk add --no-cache ca-certificates docker-cli wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gateway .

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:3000/health || exit 1

# Run the application
CMD ["./gateway"]