FROM golang:1.24.9-alpine AS builder

WORKDIR /auth-service

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY .env .env

RUN go build -o auth-service

FROM alpine:latest

WORKDIR /auth-service

COPY --from=builder /auth-service .
COPY --from=builder /auth-service/.env .env

EXPOSE 50051

CMD ["./auth-service"]