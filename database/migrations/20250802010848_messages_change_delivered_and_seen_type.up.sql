ALTER TABLE messages DROP COLUMN deliveredTimestamp;
ALTER TABLE messages DROP COLUMN seenTimestamp;

ALTER TABLE messages ADD COLUMN deliveredTimestamp TIMESTAMP WITH TIME ZONE;
ALTER TABLE messages ADD COLUMN seenTimestamp TIMESTAMP WITH TIME ZONE;