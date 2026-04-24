# ---- Build Stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# ---- Run Stage ----
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations (optional — run manually via Supabase SQL Editor)
COPY --from=builder /app/migrations ./migrations

# Expose port (Render sets PORT dynamically)
EXPOSE 8080

CMD ["./server"]
