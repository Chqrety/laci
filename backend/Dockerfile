# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o laci-backend .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o laci-seed ./cmd/seed

# Stage 2: Production
FROM alpine:latest
# Instal data zona waktu & ca-certificates
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /app
RUN mkdir -p /app/uploads
COPY --from=builder /app/laci-backend .
COPY --from=builder /app/laci-seed .
EXPOSE 8080
CMD ["./laci-backend"]
