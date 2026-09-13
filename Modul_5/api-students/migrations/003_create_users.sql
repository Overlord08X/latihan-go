CREATE TABLE IF NOT EXISTS users (
    id         SERIAL          PRIMARY KEY,
    username   VARCHAR(50)     NOT NULL,
    email      VARCHAR(100)    NOT NULL,
    password   VARCHAR(255)    NOT NULL,   -- bcrypt hash, BUKAN plaintext
    role       VARCHAR(20)     NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Unique index case-insensitive pada username dan email
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_idx ON users (LOWER(username));
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_idx    ON users (LOWER(email));
