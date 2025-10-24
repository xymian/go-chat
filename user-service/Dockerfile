FROM golang:1.24.9-alpine AS builder

WORKDIR /user-service

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY database ./database
COPY .env .env

RUN go build -o user-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /user-service .
COPY --from=builder /user-service/database ./database
COPY --from=builder /user-service/.env .env

EXPOSE 50052

CMD ["./user-service"]