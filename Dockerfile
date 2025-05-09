FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o feed-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/feed-service .
# Remove the .env copy since it doesn't exist and will be provided via environment variables
# COPY .env .

RUN adduser -D appuser
USER appuser

EXPOSE 4004

CMD ["./feed-service"] 