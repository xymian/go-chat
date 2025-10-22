FROM golang:1.24.9-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o user-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/user-service .

EXPOSE 50052

CMD ["./user-service"]