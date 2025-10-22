FROM golang:1.24.9-alpine AS builder

WORKDIR /app

COPY go.sum go.mod ./

RUN go mod download

COPY . .

RUN go build -o chat-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/chat-service .

EXPOSE 50053

CMD ["./chat-service"]