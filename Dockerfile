# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Build a static binary (SQL migrations are embedded via go:embed, so they ship inside it).
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/fundly ./cmd/fundly

# ---- runtime stage ----
FROM alpine:3.20
# ca-certificates is required for TLS (Postgres SSL + Gmail SMTP STARTTLS).
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/fundly /app/fundly

# Default uploads location. Mount a persistent disk here (set UPLOADS_DIR to match) if you
# want avatars/attachments to survive restarts and deploys.
RUN mkdir -p /app/uploads && chown -R app:app /app
USER app

# Informational only — the app binds to $PORT, which the host (e.g. Render) sets.
EXPOSE 8081

ENTRYPOINT ["/app/fundly"]
