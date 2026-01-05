# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Install build tools
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate code
RUN sqlc generate
RUN templ generate

# Build with optimizations
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w" \
    -o /app/treni \
    ./cmd

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -u 1000 -G app -D app

# Copy binary and migrations
COPY --from=builder /app/treni .
COPY --from=builder /app/internal/database/migrations ./migrations

# Create data directory for SQLite
RUN mkdir -p /app/data && chown -R app:app /app

USER app

ENV ADDR=:8080
ENV DATABASE_PATH=/app/data/treni.db

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/treni"]
