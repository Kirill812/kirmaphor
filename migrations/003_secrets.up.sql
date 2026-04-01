CREATE TABLE secrets (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    type                TEXT NOT NULL CHECK (type IN ('ssh', 'login_password', 'string', 'cloud_aws', 'cloud_azure', 'cloud_gcp')),
    encrypted_value     BYTEA NOT NULL,
    nonce               BYTEA NOT NULL,
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_secrets_project ON secrets(project_id);
