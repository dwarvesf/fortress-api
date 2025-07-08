-- +migrate Up
CREATE TABLE IF NOT EXISTS agent_api_keys (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    updated_at TIMESTAMP(6) DEFAULT (NOW()),
    name       TEXT NOT NULL,           -- Agent identifier
    api_key    TEXT UNIQUE NOT NULL,    -- Hashed API key
    permissions JSONB DEFAULT '[]'::JSONB,     -- Agent-specific permissions
    rate_limit INTEGER DEFAULT 1000,   -- Requests per hour
    is_active  BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP(6)
);

-- Add index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_agent_api_keys_api_key ON agent_api_keys(api_key);
CREATE INDEX IF NOT EXISTS idx_agent_api_keys_active ON agent_api_keys(is_active) WHERE is_active = TRUE;

-- +migrate Down
DROP INDEX IF EXISTS idx_agent_api_keys_active;
DROP INDEX IF EXISTS idx_agent_api_keys_api_key;
DROP TABLE IF EXISTS agent_api_keys;