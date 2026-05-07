module github.com/te6lim/go-chat/user-service

go 1.24.0

toolchain go1.24.9

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/golang-migrate/migrate/v4 v4.19.0
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.7.6
	github.com/joho/godotenv v1.5.1
	github.com/te6lim/go-chat-protos v0.0.0-20251209134524-fe64358ff9b7
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

replace github.com/te6lim/go-chat-protos => /Users/te6lim/go-development/go-chat-protos

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)
