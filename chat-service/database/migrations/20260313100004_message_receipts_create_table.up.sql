CREATE TABLE IF NOT EXISTS message_receipts (
    id               SERIAL PRIMARY KEY,
    messageReference UUID         NOT NULL,
    username         VARCHAR(50)  NOT NULL,
    deliveredAt      TIMESTAMP WITH TIME ZONE,
    seenAt           TIMESTAMP WITH TIME ZONE,
    createdAt        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (messageReference, username)
);
