FROM golang:1.24.9-alpine AS builder

WORKDIR /user-service

COPY go-chat-protos /go-chat-protos
COPY go-chat/user-service/go.mod go-chat/user-service/go.sum ./

RUN sed -i 's|=> .*go-chat-protos|=> /go-chat-protos|' go.mod && go mod download

COPY go-chat/user-service .

COPY go-chat/user-service/database ./database
COPY go-chat/user-service/.env .env

RUN sed -i 's|=> .*go-chat-protos|=> /go-chat-protos|' go.mod && go build -o user-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /user-service .
COPY --from=builder /user-service/database ./database
COPY --from=builder /user-service/.env .env

EXPOSE 50052

CMD ["./user-service"]
