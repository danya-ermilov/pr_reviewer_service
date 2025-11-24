# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o prreview ./cmd/prreview

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates bash postgresql-client
COPY --from=builder /app/prreview .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/swagger-ui ./swagger-ui
COPY entrypoint.sh ./entrypoint.sh
EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
