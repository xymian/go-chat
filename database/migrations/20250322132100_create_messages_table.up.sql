CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    messageReference TEXT NOT NULL,
    text TEXT NOT NULL,
    sender TEXT NOT NULL,
    receiver TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    chatReference REFERENCES chats (chatReference),
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);