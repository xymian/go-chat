FROM golang:1.24.9-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o auth-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth-service .

EXPOSE 50051

CMD ["./auth-service"]