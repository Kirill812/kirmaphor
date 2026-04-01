-- migrations/004_execution.up.sql

-- Inventories: define target hosts (created before job_templates since templates reference them)
CREATE TABLE inventories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    type            TEXT NOT NULL CHECK (type IN ('static', 'static-yaml', 'file', 'aws-ec2', 'azure-vmss', 'gcp-gce')),
    inventory_data  TEXT,                   -- inline static inventory content (for static/static-yaml)
    repository_id   UUID,                   -- set after repositories table created
    inventory_path  TEXT,                   -- path in repo (for file type)
    ssh_key_id      UUID REFERENCES secrets(id) ON DELETE SET NULL,
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Repositories: git repos containing playbooks
CREATE TABLE repositories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    git_url         TEXT NOT NULL,
    git_branch      TEXT NOT NULL DEFAULT 'main',
    ssh_key_id      UUID REFERENCES secrets(id) ON DELETE SET NULL,
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Add FK from inventories to repositories now that repositories exists
ALTER TABLE inventories ADD CONSTRAINT fk_inventories_repository
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE SET NULL;

-- Job templates: define what to run
CREATE TABLE job_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    playbook        TEXT NOT NULL,          -- path relative to repo root
    inventory_id    UUID REFERENCES inventories(id) ON DELETE SET NULL,
    repository_id   UUID NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
    environment     JSONB NOT NULL DEFAULT '{}',
    arguments       TEXT NOT NULL DEFAULT '',
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Tasks: individual execution records
CREATE TYPE task_status AS ENUM ('waiting', 'running', 'success', 'error', 'stopped');

CREATE TABLE tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    template_id     UUID NOT NULL REFERENCES job_templates(id) ON DELETE RESTRICT,
    status          task_status NOT NULL DEFAULT 'waiting',
    message         TEXT NOT NULL DEFAULT '',
    playbook        TEXT NOT NULL,
    inventory_id    UUID REFERENCES inventories(id) ON DELETE SET NULL,
    repository_id   UUID NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
    git_branch      TEXT NOT NULL,
    commit_hash     TEXT,
    arguments       TEXT NOT NULL DEFAULT '',
    environment     JSONB NOT NULL DEFAULT '{}',
    created_by      UUID NOT NULL REFERENCES users(id),
    schedule_id     UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ
);

-- Task logs: output lines from ansible-playbook
CREATE TABLE task_logs (
    id              BIGSERIAL PRIMARY KEY,
    task_id         UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    output          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Schedules: cron and one-shot triggers
CREATE TYPE schedule_type AS ENUM ('cron', 'run_at');

CREATE TABLE schedules (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    template_id         UUID NOT NULL REFERENCES job_templates(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    type                schedule_type NOT NULL DEFAULT 'cron',
    cron_format         TEXT,
    run_at              TIMESTAMPTZ,
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    delete_after_run    BOOLEAN NOT NULL DEFAULT FALSE,
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_run_at         TIMESTAMPTZ,
    UNIQUE(project_id, name),
    CHECK (
        (type = 'cron'   AND cron_format IS NOT NULL) OR
        (type = 'run_at' AND run_at IS NOT NULL)
    )
);

-- Add FK from tasks to schedules now that schedules exists
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_schedule
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE SET NULL;

CREATE INDEX idx_tasks_project        ON tasks(project_id);
CREATE INDEX idx_tasks_template       ON tasks(template_id);
CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX idx_task_logs_task       ON task_logs(task_id);
CREATE INDEX idx_schedules_project    ON schedules(project_id);
CREATE INDEX idx_inventories_project  ON inventories(project_id);
CREATE INDEX idx_repositories_project ON repositories(project_id);
CREATE INDEX idx_templates_project    ON job_templates(project_id);
