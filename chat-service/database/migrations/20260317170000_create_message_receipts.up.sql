CREATE TABLE IF NOT EXISTS message_receipts (
    id SERIAL PRIMARY KEY,
    messageReference UUID NOT NULL,
    username VARCHAR(50) NOT NULL,
    deliveredTimestamp TIMESTAMP WITH TIME ZONE,
    seenTimestamp TIMESTAMP WITH TIME ZONE,
    UNIQUE(messageReference, username),
    FOREIGN KEY (messageReference) REFERENCES messages(messageReference) ON DELETE CASCADE
);

ALTER TABLE messages DROP COLUMN deliveredTimestamp;
ALTER TABLE messages DROP COLUMN seenTimestamp;
