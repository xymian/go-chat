ALTER TABLE messages ADD COLUMN textMessage TEXT;
ALTER TABLE messages RENAME COLUMN timestamp TO messageTimestamp;