# Go-Chat

**Go-Chat** is a personal project of mine, where I explore webSocket technology, API development, databases, and how the client-side interacts with them. The goal of this project is understand how real-time messaging works, while building one that works.

## Features

- **Multiple private one-on-one messaging**: Users can send messages privately to other users.✅ (TESTABLE VERSION DONE)
- **Group chat**: Users can send messages to multiple users in a group chat.⏳🔜 (IN PROGRESS)
- **Authentication and login**: Users can securely log in to the chat server.⏳🔜 (IN PROGRESS)
- **Chat history**: The server stores chat history for each user, allowing them to view past messages.✅ (TESTABLE VERSION DONE)
- **Front-end Client**: (Currently CLI) A user-friendly application for interacting with the chat server⏳🔜 (IN PROGRESS)

## Getting Started

To get started with **Go-Chat**, follow the instructions below to set up your local environment.

### Prerequisites

Make sure you have the following installed:

- [Go](https://golang.org/dl/) (version 1.18+ recommended)
- A text editor of your choice (e.g., Visual Studio Code, Sublime Text)
- A terminal or command prompt

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/te6lim/go-chat.git
   cd go-chat
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Run the server:

   ```bash
   go run client.go
   ```

4. For now, the server will run on `http://localhost:8080` (or the port you have configured).

### Usage

Once the server is running, open your web browser and navigate to the server's URL to test the application

## Contributing

Feel free to open issues or submit pull requests if you'd like to contribute

## TODO

- Use postgres for DB implementation
