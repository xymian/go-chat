UPDATE messages SET receiverUsername = '' WHERE receiverUsername IS NULL;
ALTER TABLE messages ALTER COLUMN receiverUsername SET NOT NULL;
