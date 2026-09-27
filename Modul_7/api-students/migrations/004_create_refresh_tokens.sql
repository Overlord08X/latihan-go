CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         SERIAL          PRIMARY KEY,
    user_id    INT             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64)     NOT NULL UNIQUE,  -- SHA-256 hex dari refresh token
    is_revoked BOOLEAN         NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ     NOT NULL,
    created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx    ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_token_hash_idx ON refresh_tokens(token_hash);
