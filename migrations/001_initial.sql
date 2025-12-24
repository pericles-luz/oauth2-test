-- migrations/001_initial.sql
-- Initial database schema for OAuth2 test tool

CREATE TABLE IF NOT EXISTS http_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_method TEXT NOT NULL,
    request_url TEXT NOT NULL,
    request_headers TEXT,              -- JSON serialized headers
    request_body TEXT,
    response_status INTEGER,
    response_headers TEXT,             -- JSON serialized headers
    response_body TEXT,
    duration_ms INTEGER,
    endpoint_type TEXT,                -- authorize/token/userinfo/refresh/revoke/jwks/discovery
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_history_created ON http_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_endpoint ON http_history(endpoint_type);
