# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Start all services with their databases
docker-compose up --build

# Build a single service
cd <service-name> && go build -o <service-name>

# Run a single service locally (requires .env in service root)
cd <service-name> && go run main.go
```

There are no automated tests. Validation is done by running the services and testing endpoints manually or via the [Android client](https://github.com/te6lim/go-chat.mobile.android).

## Architecture

Four independent Go microservices, each with its own `go.mod`:

| Service | Port | Role |
|---|---|---|
| `auth-service` | `:50051` | JWT token generation, register/login/logout |
| `user-service` | `:50052` | User profiles; also exposes a gRPC server consumed by auth and chat |
| `chat-service` | `:50053` | WebSocket rooms, message persistence, message receipts |
| `storage-service` | `:50054` | Avatar uploads via streaming gRPC (no HTTP routes) |

**Databases** (separate Postgres instances):
- `go_chat` (port 5433): chats, messages, participants, receipts — owned by chat-service
- `go_chat_user` (port 5434): user profiles, conversations — owned by user-service

**gRPC call graph**: auth-service → user-service; chat-service → user-service → storage-service. Proto definitions live in the external repo `github.com/te6lim/go-chat-protos`.

### Dual WebSocket design in chat-service

Each connected client holds up to two WebSocket connections simultaneously:

1. **Room socket** (`/room/{chatReference}`) — for sending/receiving messages in a specific chat. User is `ONLINE`.
2. **Conversations socket** (`/conversations/{username}`) — always-open socket for receiving cross-chat notifications (new messages, invite events) while the user is not in a specific room. User is `AWAY`.

Presence state (`ONLINE`, `AWAY`, `OFFLINE`) is managed by `ListenForActiveUsers()` in `chat/socket-user.go`, which serializes all transitions over channels to avoid races. The `activeSocketUsers` map uses an `RWMutex` for concurrent reads.

When a room socket disconnects, the user transitions to `AWAY` (not removed) if their conversations socket is still alive. A stale-disconnect guard compares the stored `PrivateConn` pointer before applying cleanup.

### Message lifecycle

1. Client sends `database.Message` JSON over WebSocket.
2. `MaybeInsertAndReturnMostUpToDateMessage` either inserts a new message or returns the existing one and updates its delivery/seen receipt.
3. `Room.Run()` forwards the message to all online participants; AWAY participants receive it via `Socketuser.Notify`.
4. Messages are purged hourly once fully acknowledged (`deliveredTimestamp` + `seenTimestamp` set) unless `is_backed_up` is true.

### Invite flow

Private chats start with an `INVITE_PENDING` state. The invitee receives a `CHAT_INVITE` message status over their conversations socket. They respond `ACCEPT_INVITE` or `DECLINE_INVITE`. Only after acceptance does the room socket endpoint (`/room/{chatRef}`) get registered. Revoked invites are persisted in `revoked_invites` and replayed on conversations socket reconnect.

## Critical Patterns

### Response model
Every HTTP handler returns `models.Response[T]` — defined identically in each service's `models/models.go`:
```go
type Response[T any] struct {
    Data        *T     `json:"data"`
    Message     string `json:"message"`
    Error       string `json:"error"`
    StatusCode  int    `json:"statusCode"`
    IsSuccessful bool  `json:"isSuccessful"`
}
```

### JWT middleware
`middleware.WithJWTMiddleware(handler)` validates `Authorization: Bearer <token>` using the shared `JWT_SECRET` env var and injects the username into the request context via `ContextKeyUsername`. Required on all protected routes.

### No ORM
Raw SQL only via `database.Instance` (`*sql.DB`). Queries live in `database/*.go`. Migrations auto-run at startup from `database/migrations/` (golang-migrate format, `YYYYMMDDHHMMSS_description.{up,down}.sql`).

### gRPC clients
Each service stores its gRPC client in a package-level variable (e.g., `service.UserService`). In Docker, services address each other by hostname (e.g., `"user-service:50052"`), not `localhost`.

## Environment Variables

Each service reads a `.env` file in its own directory. All services share:
- `JWT_SECRET` — must be identical across all services
- `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `SSL_MODE` (chat-service and user-service only)

## Constraints

- **Ports are fixed**: 50051–50054 — do not change in code.
- **Proto changes require updating `go-chat-protos`** (separate repo) and re-importing here.
- **Docker hostnames**: inter-service calls use container names, not `localhost`.
