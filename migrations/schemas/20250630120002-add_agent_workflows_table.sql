-- +migrate Up
CREATE TABLE IF NOT EXISTS agent_workflows (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    updated_at TIMESTAMP(6) DEFAULT (NOW()),
    workflow_type TEXT NOT NULL,       -- staff_project, process_payroll, etc.
    status     TEXT NOT NULL,          -- pending, in_progress, completed, failed
    input_data JSONB NOT NULL,
    output_data JSONB,
    steps_completed INTEGER DEFAULT 0,
    total_steps INTEGER,
    agent_key_id uuid REFERENCES agent_api_keys(id),
    error_message TEXT
);

-- Add indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_agent_workflows_agent_key_id ON agent_workflows(agent_key_id);
CREATE INDEX IF NOT EXISTS idx_agent_workflows_status ON agent_workflows(status);
CREATE INDEX IF NOT EXISTS idx_agent_workflows_workflow_type ON agent_workflows(workflow_type);
CREATE INDEX IF NOT EXISTS idx_agent_workflows_created_at ON agent_workflows(created_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_agent_workflows_created_at;
DROP INDEX IF EXISTS idx_agent_workflows_workflow_type;
DROP INDEX IF EXISTS idx_agent_workflows_status;
DROP INDEX IF EXISTS idx_agent_workflows_agent_key_id;
DROP TABLE IF EXISTS agent_workflows;