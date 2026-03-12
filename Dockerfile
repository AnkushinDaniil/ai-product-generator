# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies (gcc, musl-dev for CGO/sqlite3)
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build with CGO enabled for sqlite3
ARG VERSION=dev
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-X main.version=${VERSION} -linkmode external -extldflags '-static'" \
    -o /memex ./cmd/memex

# Runtime stage
FROM alpine:3.23

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Copy binary
COPY --from=builder /memex /app/memex

# Non-root user
RUN adduser -D -g '' appuser && \
    mkdir -p /app/data && \
    chown -R appuser:appuser /app
USER appuser

# Default database path
ENV MEMEX_DB_PATH=/app/data/memex.db
ENV MEMEX_MODE=production

# MCP servers communicate via stdio (no ports exposed)
ENTRYPOINT ["/app/memex"]
