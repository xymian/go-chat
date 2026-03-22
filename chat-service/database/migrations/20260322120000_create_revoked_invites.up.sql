CREATE TABLE IF NOT EXISTS revoked_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inviteeUsername TEXT NOT NULL,
    initiatorUsername TEXT NOT NULL,
    chatReference TEXT NOT NULL,
    revokedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (inviteeUsername, chatReference)
);
