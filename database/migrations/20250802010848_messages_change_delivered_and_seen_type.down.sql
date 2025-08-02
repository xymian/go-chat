ALTER TABLE messages DROP COLUMN deliveredTimestamp;
ALTER TABLE messages DROP COLUMN seenTimestamp;

ALTER TABLE messages ADD COLUMN deliveredTimestamp TEXT;
ALTER TABLE messages ADD COLUMN seenTimestamp TEXT;