CREATE TABLE IF NOT EXISTS participants(
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    chatReference UUID NOT NULL,
    createdAt TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);