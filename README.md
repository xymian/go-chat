# Go-Chat

**Go-Chat** An exploration of WebSocket technology, API development, database design, Microservices, and all the challenges of building software at scale.

## Current Features

- **Message Receipts**: Suppports message receipts such as: SENT, DELIVERED, READ ✅
- **Ephemeral storage**: messages are only stored until they are both delivered and read. a new `is_backed_up` flag prevents deletion when set; this will be used by future backup features.
- **Multiple private one-on-one messaging**: Users can send messages privately to other users.✅
- **Group chat**: Users can send messages to multiple users in a group chat.⏳🔜 (In progress).
- **Authentication**: Users can securely log in to the chat server.✅
- **JWT-based**: route protection: For securing routes. ✅
- **Chat history**: The server stores chat history for each user, allowing them to view past messages.✅

## Tools:

- Websocket
- GRPC
- Microservices
- Docker

## Usage:

For a demonstration of how it works, you can run this server locally, and check out my [Android project](https://github.com/te6lim/go-chat.mobile.android) repo that uses this server.
