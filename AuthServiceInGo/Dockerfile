# ─── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a statically-linked binary — no libc needed in the final image
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /gateway main.go

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM alpine:3.21

# ca-certificates — needed for HTTPS calls to upstream services
# tzdata         — needed for correct time zone handling in JWT expiry
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /gateway .

EXPOSE 8080
CMD ["./gateway"]
