CREATE TABLE platforms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    api_endpoint TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE social_accounts (
    id SERIAL PRIMARY KEY,
    username VARCHAR NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    platform_id INT NOT NULL REFERENCES platforms(id) ON DELETE CASCADE,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);