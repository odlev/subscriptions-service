FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main /app/main

COPY .env /app/.env
COPY config.yaml /app/config.yaml

CMD [ "/app/main" ]
