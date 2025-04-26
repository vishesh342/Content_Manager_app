CREATE TABLE posts (
    id VARCHAR(16) PRIMARY KEY,
    content TEXT NOT NULL,
    media_type VARCHAR NOT NULL CHECK (media_type IN ('NONE', 'IMAGE', 'VIDEO', 'DOCUMENT', 'ARTICLE', 'MULTIPLE_MEDIA')),
    media_urns JSONB,
    scheduled_time TIMESTAMPTZ,
    visibility VARCHAR NOT NULL CHECK (visibility IN ('PUBLIC', 'CONNECTIONS')),
    account_id VARCHAR NOT NULL REFERENCES social_accounts(platform_username),
    created_at TIMESTAMPTZ DEFAULT NOW()
);