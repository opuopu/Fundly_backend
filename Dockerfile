FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /fundly ./cmd/fundly

# Production runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -S fundly && adduser -S fundly -G fundly

WORKDIR /app

# Create uploads directory and assign ownership
RUN mkdir -p /app/uploads && chown -R fundly:fundly /app

COPY --from=builder /fundly /app/fundly

USER fundly

ENV PORT=8081 \
    APP_ENV=production \
    UPLOADS_DIR=/app/uploads

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:${PORT}/health || exit 1

CMD ["/app/fundly"]
