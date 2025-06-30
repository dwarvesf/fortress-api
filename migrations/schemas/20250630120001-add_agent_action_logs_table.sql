-- +migrate Up
CREATE TABLE IF NOT EXISTS agent_action_logs (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    agent_key_id uuid REFERENCES agent_api_keys(id),
    tool_name  TEXT NOT NULL,
    input_data JSONB,
    output_data JSONB,
    status     TEXT NOT NULL,           -- success, error, timeout
    duration_ms INTEGER,
    error_message TEXT
);

-- Add indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_agent_action_logs_agent_key_id ON agent_action_logs(agent_key_id);
CREATE INDEX IF NOT EXISTS idx_agent_action_logs_created_at ON agent_action_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_agent_action_logs_tool_name ON agent_action_logs(tool_name);
CREATE INDEX IF NOT EXISTS idx_agent_action_logs_status ON agent_action_logs(status);

-- +migrate Down
DROP INDEX IF EXISTS idx_agent_action_logs_status;
DROP INDEX IF EXISTS idx_agent_action_logs_tool_name;
DROP INDEX IF EXISTS idx_agent_action_logs_created_at;
DROP INDEX IF EXISTS idx_agent_action_logs_agent_key_id;
DROP TABLE IF EXISTS agent_action_logs;