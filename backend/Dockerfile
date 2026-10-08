# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o laci-backend .

# Stage 2: Production
FROM alpine:latest
# Instal data zona waktu agar Alpine mengenali Asia/Jakarta
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /app/laci-backend .
EXPOSE 8080
CMD ["./laci-backend"]