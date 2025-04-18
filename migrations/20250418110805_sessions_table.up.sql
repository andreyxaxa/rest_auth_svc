CREATE TABLE sessions 
(
    id VARCHAR PRIMARY KEY NOT NULL,
    user_email VARCHAR NOT NULL,
    refresh_token_hash VARCHAR NOT NULL,
    is_revoked bool NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);