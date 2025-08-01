ALTER TABLE messages DROP COLUMN deliveredTimestamp;
ALTER TABLE messages DROP COLUMN seenTimestamp;

ALTER TABLE messages ADD COLUMN delivered BOOLEAN;
ALTER TABLE messages ADD COLUMN seen BOOLEAN;