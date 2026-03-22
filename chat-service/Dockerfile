FROM golang:1.24.9-alpine AS builder

WORKDIR /chat-service

COPY go-chat-protos /go-chat-protos
COPY go-chat/chat-service/go.sum go-chat/chat-service/go.mod ./

RUN go mod download

COPY go-chat/chat-service .
COPY go-chat/chat-service/.env .env
COPY go-chat/chat-service/database ./database

RUN go build -o chat-service

FROM alpine:latest

WORKDIR /chat-service

COPY --from=builder /chat-service .
COPY --from=builder /chat-service/.env .env
COPY --from=builder /chat-service/database ./database

EXPOSE 50053

CMD ["./chat-service"]
