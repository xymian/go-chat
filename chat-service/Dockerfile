FROM golang:1.24.9-alpine AS builder

WORKDIR /chat-service

COPY go.sum go.mod ./

RUN go mod download

COPY . .
COPY .env .env
COPY database ./database

RUN go build -o chat-service

FROM alpine:latest

WORKDIR /chat-service

COPY --from=builder /chat-service .
COPY --from=builder /chat-service/.env .env
COPY --from=builder /chat-service/database ./database

EXPOSE 50053

CMD ["./chat-service"]