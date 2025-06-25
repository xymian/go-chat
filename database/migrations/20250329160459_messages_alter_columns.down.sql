ALTER TABLE messages DROP COLUMN textMessage;
ALTER TABLE messages RENAME COLUMN messageTimestamp TO timestamp;