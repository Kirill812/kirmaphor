# Kirmaphore Plan 2: Execution Core

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Ansible execution engine: job templates, task queue (TaskPool), Git integration, inventory loading, process execution, batched log streaming via WebSocket, and cron scheduling.

**Architecture:** Follows Semaphore's battle-tested patterns: channel-based TaskPool with buffered `register`/`logger` channels, a Job interface abstracted over LocalJob (subprocess), and a TaskRunner that assembles all execution context (template + inventory + repository + secrets). Logs are flushed to DB in 500ms batches. WebSocket streams live output. Scheduler polls due schedules every 30s.

**Tech Stack:** Go (existing), gorilla/websocket, robfig/cron v3, exec.Command (ansible-playbook subprocess), pgx/v5, chi router (existing).

---

## File Map

```
migrations/
  004_execution.up.sql          -- templates, tasks, task_logs, inventories,
                                --   repositories, runners, schedules
  004_execution.down.sql

internal/db/models/
  template.go                   -- JobTemplate struct
  task.go                       -- Task struct (execution record)
  task_log.go                   -- TaskLog struct (single output line)
  inventory.go                  -- Inventory struct
  repository.go                 -- Repository struct
  schedule.go                   -- Schedule struct

internal/db/queries/
  template.go                   -- CreateTemplate, GetTemplate, ListTemplates, DeleteTemplate
  task.go                       -- CreateTask, GetTask, UpdateTaskStatus, ListTasks
  task_log.go                   -- AppendLogs (batch insert), GetLogs
  inventory.go                  -- CreateInventory, GetInventory, ListInventories, DeleteInventory
  repository.go                 -- CreateRepository, GetRepository, ListRepositories, DeleteRepository
  schedule.go                   -- CreateSchedule, GetSchedule, ListSchedules,
                                --   UpdateSchedule, GetDueSchedules, TouchSchedule

internal/execution/
  job.go                        -- Job interface + Status constants
  local_job.go                  -- LocalJob: runs ansible-playbook subprocess
  log_writer.go                 -- LogWriter: batched DB flush every 500ms
  task_runner.go                -- TaskRunner: assembles context, runs job, stores result
  task_pool.go                  -- TaskPool: channel queue, goroutine manager

internal/git/
  gateway.go                    -- CloneOrPull(repo, keyMaterial) (dir, error)
  ssh_auth.go                   -- WriteKeyFile(keyPEM) (path, cleanup, error)

internal/inventory/
  loader.go                     -- Load(inv Inventory, secrets map) (tmpFilePath, cleanup, error)

internal/scheduler/
  scheduler.go                  -- Scheduler: polls DB every 30s, enqueues due schedules

internal/api/handlers/
  templates.go                  -- CRUD /projects/:projectId/templates
  tasks.go                      -- POST /run, GET /tasks/:id, GET /projects/:id/tasks
  inventories.go                -- CRUD /projects/:projectId/inventories
  repositories.go               -- CRUD /projects/:projectId/repositories
  schedules.go                  -- CRUD /projects/:projectId/schedules
  logs.go                       -- GET /tasks/:id/logs (JSON) + WS /tasks/:id/logs/stream

internal/api/router.go          -- add execution routes (Modify)
```

---

## Task 1: DB Migration — Execution Tables

**Files:**
- Create: `migrations/004_execution.up.sql`
- Create: `migrations/004_execution.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- migrations/004_execution.up.sql

-- Job templates: define what to run
CREATE TABLE job_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    playbook        TEXT NOT NULL,          -- path relative to repo root
    inventory_id    UUID REFERENCES inventories(id) ON DELETE SET NULL,
    repository_id   UUID NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
    environment     JSONB NOT NULL DEFAULT '{}',  -- extra env vars
    arguments       TEXT NOT NULL DEFAULT '',      -- extra ansible-playbook args
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Inventories: define target hosts
CREATE TABLE inventories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    type            TEXT NOT NULL CHECK (type IN ('static', 'static-yaml', 'file', 'aws-ec2', 'azure-vmss', 'gcp-gce')),
    inventory_data  TEXT,                   -- inline static inventory content (for static/static-yaml)
    repository_id   UUID REFERENCES repositories(id) ON DELETE SET NULL,
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

-- Tasks: individual execution records
CREATE TYPE task_status AS ENUM ('waiting', 'running', 'success', 'error', 'stopped');

CREATE TABLE tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    template_id     UUID NOT NULL REFERENCES job_templates(id) ON DELETE RESTRICT,
    status          task_status NOT NULL DEFAULT 'waiting',
    message         TEXT NOT NULL DEFAULT '',   -- error message or short status
    playbook        TEXT NOT NULL,              -- snapshot from template at run time
    inventory_id    UUID REFERENCES inventories(id) ON DELETE SET NULL,
    repository_id   UUID NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
    git_branch      TEXT NOT NULL,              -- branch used for this run
    commit_hash     TEXT,                       -- populated after checkout
    arguments       TEXT NOT NULL DEFAULT '',
    environment     JSONB NOT NULL DEFAULT '{}',
    created_by      UUID NOT NULL REFERENCES users(id),
    schedule_id     UUID,                       -- null if triggered manually
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
    cron_format         TEXT,               -- required when type='cron'
    run_at              TIMESTAMPTZ,        -- required when type='run_at'
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    delete_after_run    BOOLEAN NOT NULL DEFAULT FALSE,
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_run_at         TIMESTAMPTZ,
    UNIQUE(project_id, name)
);

CREATE INDEX idx_tasks_project        ON tasks(project_id);
CREATE INDEX idx_tasks_template       ON tasks(template_id);
CREATE INDEX idx_tasks_status         ON tasks(status);
CREATE INDEX idx_task_logs_task       ON task_logs(task_id);
CREATE INDEX idx_schedules_project    ON schedules(project_id);
CREATE INDEX idx_inventories_project  ON inventories(project_id);
CREATE INDEX idx_repositories_project ON repositories(project_id);
CREATE INDEX idx_templates_project    ON job_templates(project_id);
```

- [ ] **Step 2: Write down migration**

```sql
-- migrations/004_execution.down.sql
DROP INDEX IF EXISTS idx_templates_project;
DROP INDEX IF EXISTS idx_repositories_project;
DROP INDEX IF EXISTS idx_inventories_project;
DROP INDEX IF EXISTS idx_schedules_project;
DROP INDEX IF EXISTS idx_task_logs_task;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_template;
DROP INDEX IF EXISTS idx_tasks_project;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS task_logs;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS job_templates;
DROP TABLE IF EXISTS inventories;
DROP TABLE IF EXISTS repositories;
DROP TYPE IF EXISTS schedule_type;
DROP TYPE IF EXISTS task_status;
```

- [ ] **Step 3: Apply migration**

```bash
cd /Users/kgory/dev/kirmaphor
export DATABASE_URL="postgres://kirmaphore:kirmaphore@localhost:5432/kirmaphore?sslmode=disable"
export MASTER_KEY="0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
go run cmd/migrate/main.go migrations
```

Expected: no error. Verify:
```bash
docker exec kirmaphore-db psql -U kirmaphore -c "\dt" | grep -E "tasks|templates|inventories|repositories|schedules"
```

Expected: all 6 new tables listed.

- [ ] **Step 4: Commit**

```bash
git add migrations/004_execution.up.sql migrations/004_execution.down.sql
git commit -m "feat: execution tables migration (templates, tasks, inventories, repos, schedules)"
```

---

## Task 2: DB Models — Execution Entities

**Files:**
- Create: `internal/db/models/template.go`
- Create: `internal/db/models/task.go`
- Create: `internal/db/models/task_log.go`
- Create: `internal/db/models/inventory.go`
- Create: `internal/db/models/repository.go`
- Create: `internal/db/models/schedule.go`

- [ ] **Step 1: Write all models**

```go
// internal/db/models/template.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type JobTemplate struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	Name         string
	Description  string
	Playbook     string
	InventoryID  *uuid.UUID
	RepositoryID uuid.UUID
	Environment  map[string]string
	Arguments    string
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
}
```

```go
// internal/db/models/task.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusWaiting TaskStatus = "waiting"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusError   TaskStatus = "error"
	TaskStatusStopped TaskStatus = "stopped"
)

type Task struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	TemplateID   uuid.UUID
	Status       TaskStatus
	Message      string
	Playbook     string
	InventoryID  *uuid.UUID
	RepositoryID uuid.UUID
	GitBranch    string
	CommitHash   *string
	Arguments    string
	Environment  map[string]string
	CreatedBy    uuid.UUID
	ScheduleID   *uuid.UUID
	CreatedAt    time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
}
```

```go
// internal/db/models/task_log.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type TaskLog struct {
	ID        int64
	TaskID    uuid.UUID
	Output    string
	CreatedAt time.Time
}
```

```go
// internal/db/models/inventory.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type InventoryType string

const (
	InventoryTypeStatic     InventoryType = "static"
	InventoryTypeStaticYAML InventoryType = "static-yaml"
	InventoryTypeFile       InventoryType = "file"
	InventoryTypeAWSEC2     InventoryType = "aws-ec2"
	InventoryTypeAzureVMSS  InventoryType = "azure-vmss"
	InventoryTypeGCPGCE     InventoryType = "gcp-gce"
)

type Inventory struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	Name          string
	Type          InventoryType
	InventoryData *string   // inline content for static/static-yaml
	RepositoryID  *uuid.UUID
	InventoryPath *string   // path in repo for file type
	SSHKeyID      *uuid.UUID
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
}
```

```go
// internal/db/models/repository.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type Repository struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	Name       string
	GitURL     string
	GitBranch  string
	SSHKeyID   *uuid.UUID
	CreatedBy  uuid.UUID
	CreatedAt  time.Time
}
```

```go
// internal/db/models/schedule.go
package models

import (
	"time"
	"github.com/google/uuid"
)

type ScheduleType string

const (
	ScheduleTypeCron  ScheduleType = "cron"
	ScheduleTypeRunAt ScheduleType = "run_at"
)

type Schedule struct {
	ID              uuid.UUID
	ProjectID       uuid.UUID
	TemplateID      uuid.UUID
	Name            string
	Type            ScheduleType
	CronFormat      *string
	RunAt           *time.Time
	Active          bool
	DeleteAfterRun  bool
	CreatedBy       uuid.UUID
	CreatedAt       time.Time
	LastRunAt       *time.Time
}
```

- [ ] **Step 2: Build check**

```bash
cd /Users/kgory/dev/kirmaphor
go build ./internal/db/models/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/db/models/
git commit -m "feat: execution entity models (template, task, inventory, repository, schedule)"
```

---

## Task 3: DB Queries — Execution CRUD

**Files:**
- Create: `internal/db/queries/template.go`
- Create: `internal/db/queries/task.go`
- Create: `internal/db/queries/task_log.go`
- Create: `internal/db/queries/inventory.go`
- Create: `internal/db/queries/repository.go`
- Create: `internal/db/queries/schedule.go`

- [ ] **Step 1: Write template queries**

```go
// internal/db/queries/template.go
package queries

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateTemplate(ctx context.Context, pool *pgxpool.Pool, t *models.JobTemplate) (*models.JobTemplate, error) {
	result := &models.JobTemplate{}
	err := pool.QueryRow(ctx,
		`INSERT INTO job_templates
		   (project_id, name, description, playbook, inventory_id, repository_id,
		    environment, arguments, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id, project_id, name, description, playbook, inventory_id,
		           repository_id, environment, arguments, created_by, created_at`,
		t.ProjectID, t.Name, t.Description, t.Playbook, t.InventoryID, t.RepositoryID,
		t.Environment, t.Arguments, t.CreatedBy,
	).Scan(&result.ID, &result.ProjectID, &result.Name, &result.Description,
		&result.Playbook, &result.InventoryID, &result.RepositoryID,
		&result.Environment, &result.Arguments, &result.CreatedBy, &result.CreatedAt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetTemplate(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.JobTemplate, error) {
	t := &models.JobTemplate{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, name, description, playbook, inventory_id,
		        repository_id, environment, arguments, created_by, created_at
		 FROM job_templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.Playbook,
		&t.InventoryID, &t.RepositoryID, &t.Environment, &t.Arguments,
		&t.CreatedBy, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func ListTemplates(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) ([]*models.JobTemplate, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, name, description, playbook, inventory_id,
		        repository_id, environment, arguments, created_by, created_at
		 FROM job_templates WHERE project_id = $1 ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var templates []*models.JobTemplate
	for rows.Next() {
		t := &models.JobTemplate{}
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.Playbook,
			&t.InventoryID, &t.RepositoryID, &t.Environment, &t.Arguments,
			&t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func DeleteTemplate(ctx context.Context, pool *pgxpool.Pool, id, projectID uuid.UUID) error {
	tag, err := pool.Exec(ctx,
		`DELETE FROM job_templates WHERE id = $1 AND project_id = $2`, id, projectID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound // reuse sentinel — "not found"
	}
	return nil
}
```

- [ ] **Step 2: Write task queries**

```go
// internal/db/queries/task.go
package queries

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateTask(ctx context.Context, pool *pgxpool.Pool, t *models.Task) (*models.Task, error) {
	result := &models.Task{}
	err := pool.QueryRow(ctx,
		`INSERT INTO tasks
		   (project_id, template_id, status, playbook, inventory_id,
		    repository_id, git_branch, arguments, environment, created_by, schedule_id)
		 VALUES ($1,$2,'waiting',$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING id, project_id, template_id, status, message, playbook, inventory_id,
		           repository_id, git_branch, commit_hash, arguments, environment,
		           created_by, schedule_id, created_at, started_at, finished_at`,
		t.ProjectID, t.TemplateID, t.Playbook, t.InventoryID, t.RepositoryID,
		t.GitBranch, t.Arguments, t.Environment, t.CreatedBy, t.ScheduleID,
	).Scan(&result.ID, &result.ProjectID, &result.TemplateID, &result.Status,
		&result.Message, &result.Playbook, &result.InventoryID, &result.RepositoryID,
		&result.GitBranch, &result.CommitHash, &result.Arguments, &result.Environment,
		&result.CreatedBy, &result.ScheduleID, &result.CreatedAt, &result.StartedAt, &result.FinishedAt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetTask(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Task, error) {
	t := &models.Task{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, template_id, status, message, playbook, inventory_id,
		        repository_id, git_branch, commit_hash, arguments, environment,
		        created_by, schedule_id, created_at, started_at, finished_at
		 FROM tasks WHERE id = $1`, id,
	).Scan(&t.ID, &t.ProjectID, &t.TemplateID, &t.Status, &t.Message,
		&t.Playbook, &t.InventoryID, &t.RepositoryID, &t.GitBranch, &t.CommitHash,
		&t.Arguments, &t.Environment, &t.CreatedBy, &t.ScheduleID,
		&t.CreatedAt, &t.StartedAt, &t.FinishedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func UpdateTaskStatus(ctx context.Context, pool *pgxpool.Pool,
	id uuid.UUID, status models.TaskStatus, message string) error {
	_, err := pool.Exec(ctx,
		`UPDATE tasks SET status = $1, message = $2,
		  started_at  = CASE WHEN $1 = 'running' AND started_at IS NULL THEN NOW() ELSE started_at END,
		  finished_at = CASE WHEN $1 IN ('success','error','stopped') THEN NOW() ELSE finished_at END
		 WHERE id = $3`,
		status, message, id)
	return err
}

func UpdateTaskCommit(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, hash string) error {
	_, err := pool.Exec(ctx, `UPDATE tasks SET commit_hash = $1 WHERE id = $2`, hash, id)
	return err
}

func ListTasks(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID, limit int) ([]*models.Task, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, template_id, status, message, playbook, inventory_id,
		        repository_id, git_branch, commit_hash, arguments, environment,
		        created_by, schedule_id, created_at, started_at, finished_at
		 FROM tasks WHERE project_id = $1
		 ORDER BY created_at DESC LIMIT $2`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []*models.Task
	for rows.Next() {
		t := &models.Task{}
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.TemplateID, &t.Status, &t.Message,
			&t.Playbook, &t.InventoryID, &t.RepositoryID, &t.GitBranch, &t.CommitHash,
			&t.Arguments, &t.Environment, &t.CreatedBy, &t.ScheduleID,
			&t.CreatedAt, &t.StartedAt, &t.FinishedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
```

- [ ] **Step 3: Write task_log queries**

```go
// internal/db/queries/task_log.go
package queries

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

// AppendLogs inserts a batch of output lines in a single statement.
func AppendLogs(ctx context.Context, pool *pgxpool.Pool, taskID uuid.UUID, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	batch := &pgxpool.Pool{}
	_ = batch
	// Use pgx CopyFrom for bulk insert
	rows := make([][]any, len(lines))
	for i, line := range lines {
		rows[i] = []any{taskID, line}
	}
	_, err := pool.CopyFrom(ctx,
		[]string{"task_logs"},
		[]string{"task_id", "output"},
		pgxpool.CopyFromRows(rows),
	)
	return err
}

func GetLogs(ctx context.Context, pool *pgxpool.Pool, taskID uuid.UUID) ([]*models.TaskLog, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, task_id, output, created_at
		 FROM task_logs WHERE task_id = $1 ORDER BY id ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*models.TaskLog
	for rows.Next() {
		l := &models.TaskLog{}
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Output, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func GetLogsAfter(ctx context.Context, pool *pgxpool.Pool, taskID uuid.UUID, afterID int64) ([]*models.TaskLog, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, task_id, output, created_at
		 FROM task_logs WHERE task_id = $1 AND id > $2 ORDER BY id ASC`, taskID, afterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*models.TaskLog
	for rows.Next() {
		l := &models.TaskLog{}
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Output, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
```

- [ ] **Step 4: Write inventory, repository, schedule queries**

```go
// internal/db/queries/inventory.go
package queries

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateInventory(ctx context.Context, pool *pgxpool.Pool, inv *models.Inventory) (*models.Inventory, error) {
	result := &models.Inventory{}
	err := pool.QueryRow(ctx,
		`INSERT INTO inventories
		   (project_id, name, type, inventory_data, repository_id, inventory_path, ssh_key_id, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, project_id, name, type, inventory_data, repository_id, inventory_path, ssh_key_id, created_by, created_at`,
		inv.ProjectID, inv.Name, inv.Type, inv.InventoryData, inv.RepositoryID,
		inv.InventoryPath, inv.SSHKeyID, inv.CreatedBy,
	).Scan(&result.ID, &result.ProjectID, &result.Name, &result.Type, &result.InventoryData,
		&result.RepositoryID, &result.InventoryPath, &result.SSHKeyID, &result.CreatedBy, &result.CreatedAt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetInventory(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Inventory, error) {
	inv := &models.Inventory{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, name, type, inventory_data, repository_id, inventory_path, ssh_key_id, created_by, created_at
		 FROM inventories WHERE id = $1`, id,
	).Scan(&inv.ID, &inv.ProjectID, &inv.Name, &inv.Type, &inv.InventoryData,
		&inv.RepositoryID, &inv.InventoryPath, &inv.SSHKeyID, &inv.CreatedBy, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func ListInventories(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) ([]*models.Inventory, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, name, type, inventory_data, repository_id, inventory_path, ssh_key_id, created_by, created_at
		 FROM inventories WHERE project_id = $1 ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invs []*models.Inventory
	for rows.Next() {
		inv := &models.Inventory{}
		if err := rows.Scan(&inv.ID, &inv.ProjectID, &inv.Name, &inv.Type, &inv.InventoryData,
			&inv.RepositoryID, &inv.InventoryPath, &inv.SSHKeyID, &inv.CreatedBy, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, rows.Err()
}

func DeleteInventory(ctx context.Context, pool *pgxpool.Pool, id, projectID uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM inventories WHERE id = $1 AND project_id = $2`, id, projectID)
	return err
}
```

```go
// internal/db/queries/repository.go
package queries

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateRepository(ctx context.Context, pool *pgxpool.Pool, r *models.Repository) (*models.Repository, error) {
	result := &models.Repository{}
	err := pool.QueryRow(ctx,
		`INSERT INTO repositories (project_id, name, git_url, git_branch, ssh_key_id, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, project_id, name, git_url, git_branch, ssh_key_id, created_by, created_at`,
		r.ProjectID, r.Name, r.GitURL, r.GitBranch, r.SSHKeyID, r.CreatedBy,
	).Scan(&result.ID, &result.ProjectID, &result.Name, &result.GitURL,
		&result.GitBranch, &result.SSHKeyID, &result.CreatedBy, &result.CreatedAt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetRepository(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Repository, error) {
	r := &models.Repository{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, name, git_url, git_branch, ssh_key_id, created_by, created_at
		 FROM repositories WHERE id = $1`, id,
	).Scan(&r.ID, &r.ProjectID, &r.Name, &r.GitURL, &r.GitBranch, &r.SSHKeyID, &r.CreatedBy, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func ListRepositories(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) ([]*models.Repository, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, name, git_url, git_branch, ssh_key_id, created_by, created_at
		 FROM repositories WHERE project_id = $1 ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []*models.Repository
	for rows.Next() {
		r := &models.Repository{}
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.Name, &r.GitURL, &r.GitBranch,
			&r.SSHKeyID, &r.CreatedBy, &r.CreatedAt); err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, rows.Err()
}

func DeleteRepository(ctx context.Context, pool *pgxpool.Pool, id, projectID uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM repositories WHERE id = $1 AND project_id = $2`, id, projectID)
	return err
}
```

```go
// internal/db/queries/schedule.go
package queries

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateSchedule(ctx context.Context, pool *pgxpool.Pool, s *models.Schedule) (*models.Schedule, error) {
	result := &models.Schedule{}
	err := pool.QueryRow(ctx,
		`INSERT INTO schedules
		   (project_id, template_id, name, type, cron_format, run_at, active, delete_after_run, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id, project_id, template_id, name, type, cron_format, run_at,
		           active, delete_after_run, created_by, created_at, last_run_at`,
		s.ProjectID, s.TemplateID, s.Name, s.Type, s.CronFormat, s.RunAt,
		s.Active, s.DeleteAfterRun, s.CreatedBy,
	).Scan(&result.ID, &result.ProjectID, &result.TemplateID, &result.Name, &result.Type,
		&result.CronFormat, &result.RunAt, &result.Active, &result.DeleteAfterRun,
		&result.CreatedBy, &result.CreatedAt, &result.LastRunAt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetSchedule(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Schedule, error) {
	s := &models.Schedule{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, template_id, name, type, cron_format, run_at,
		        active, delete_after_run, created_by, created_at, last_run_at
		 FROM schedules WHERE id = $1`, id,
	).Scan(&s.ID, &s.ProjectID, &s.TemplateID, &s.Name, &s.Type, &s.CronFormat,
		&s.RunAt, &s.Active, &s.DeleteAfterRun, &s.CreatedBy, &s.CreatedAt, &s.LastRunAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func ListSchedules(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) ([]*models.Schedule, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, template_id, name, type, cron_format, run_at,
		        active, delete_after_run, created_by, created_at, last_run_at
		 FROM schedules WHERE project_id = $1 ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []*models.Schedule
	for rows.Next() {
		s := &models.Schedule{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.TemplateID, &s.Name, &s.Type,
			&s.CronFormat, &s.RunAt, &s.Active, &s.DeleteAfterRun,
			&s.CreatedBy, &s.CreatedAt, &s.LastRunAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

// GetDueSchedules returns active cron schedules that need to run and run_at
// schedules whose time has passed.
func GetDueSchedules(ctx context.Context, pool *pgxpool.Pool) ([]*models.Schedule, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, template_id, name, type, cron_format, run_at,
		        active, delete_after_run, created_by, created_at, last_run_at
		 FROM schedules
		 WHERE active = TRUE
		   AND (
		     (type = 'run_at' AND run_at <= NOW())
		     OR type = 'cron'
		   )`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []*models.Schedule
	for rows.Next() {
		s := &models.Schedule{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.TemplateID, &s.Name, &s.Type,
			&s.CronFormat, &s.RunAt, &s.Active, &s.DeleteAfterRun,
			&s.CreatedBy, &s.CreatedAt, &s.LastRunAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func TouchSchedule(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, lastRunAt time.Time) error {
	_, err := pool.Exec(ctx,
		`UPDATE schedules SET last_run_at = $1 WHERE id = $2`, lastRunAt, id)
	return err
}

func DeleteSchedule(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	return err
}
```

- [ ] **Step 5: Fix AppendLogs — pgx CopyFrom syntax**

The `pgxpool.CopyFromRows` function is on `pgx`, not `pgxpool`. Replace the `AppendLogs` implementation:

```go
// internal/db/queries/task_log.go
package queries

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func AppendLogs(ctx context.Context, pool *pgxpool.Pool, taskID uuid.UUID, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	rows := make([][]any, len(lines))
	for i, line := range lines {
		rows[i] = []any{taskID, line}
	}
	_, err := pool.CopyFrom(ctx,
		pgx.Identifier{"task_logs"},
		[]string{"task_id", "output"},
		pgx.CopyFromRows(rows),
	)
	return err
}
```

- [ ] **Step 6: Build check**

```bash
cd /Users/kgory/dev/kirmaphor
go build ./internal/db/...
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add internal/db/queries/
git commit -m "feat: execution CRUD queries (templates, tasks, logs, inventories, repos, schedules)"
```

---

## Task 4: Git Gateway

**Files:**
- Create: `internal/git/gateway.go`
- Create: `internal/git/ssh_auth.go`
- Create: `internal/git/gateway_test.go`

Install dependency:
```bash
go get github.com/go-git/go-git/v5@latest
```

- [ ] **Step 1: Write failing tests**

```go
// internal/git/gateway_test.go
package git_test

import (
	"os"
	"testing"
	"github.com/kgory/kirmaphor/internal/git"
)

func TestClonePublicRepo(t *testing.T) {
	// Clone a small public repo to a temp dir
	dir, cleanup, err := git.CloneOrPull("https://github.com/nicholaswilde/hello-world-ansible.git", "main", nil)
	if err != nil {
		t.Skipf("skipping: network unavailable or repo changed: %v", err)
	}
	defer cleanup()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("expected clone dir to exist at %s", dir)
	}
	// Check that a file exists in the cloned repo
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one file in cloned repo")
	}
}

func TestWriteKeyFileCleansUp(t *testing.T) {
	keyPEM := []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nfakekey\n-----END OPENSSH PRIVATE KEY-----")
	path, cleanup, err := git.WriteKeyFile(keyPEM)
	if err != nil {
		t.Fatalf("WriteKeyFile: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("key file should exist at %s", path)
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("key file should be removed after cleanup")
	}
}
```

- [ ] **Step 2: Run, confirm FAIL**

```bash
go test ./internal/git/... -v
```

Expected: FAIL — git package undefined.

- [ ] **Step 3: Implement WriteKeyFile**

```go
// internal/git/ssh_auth.go
package git

import (
	"fmt"
	"os"
)

// WriteKeyFile writes PEM key bytes to a temp file with 0600 permissions.
// Returns (path, cleanup func, error). Caller must call cleanup() when done.
func WriteKeyFile(keyPEM []byte) (string, func(), error) {
	f, err := os.CreateTemp("", "kirmaphore-sshkey-*.pem")
	if err != nil {
		return "", nil, fmt.Errorf("create temp key file: %w", err)
	}
	if err := f.Chmod(0600); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("chmod key file: %w", err)
	}
	if _, err := f.Write(keyPEM); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("write key file: %w", err)
	}
	f.Close()
	path := f.Name()
	return path, func() { os.Remove(path) }, nil
}
```

- [ ] **Step 4: Implement CloneOrPull**

```go
// internal/git/gateway.go
package git

import (
	"fmt"
	"os"
	"os/exec"
)

// CloneOrPull clones a git repository to a temp directory, or pulls if it
// already exists. keyPEM is optional (nil for public repos / HTTPS).
// Returns (workDir, cleanup, error). Caller must call cleanup() when done.
func CloneOrPull(gitURL, branch string, keyPEM []byte) (string, func(), error) {
	workDir, err := os.MkdirTemp("", "kirmaphore-repo-*")
	if err != nil {
		return "", nil, fmt.Errorf("create work dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(workDir) }

	env := os.Environ()
	var keyCleanup func()

	if len(keyPEM) > 0 {
		keyPath, kc, err := WriteKeyFile(keyPEM)
		if err != nil {
			cleanup()
			return "", nil, err
		}
		keyCleanup = kc
		// Tell git to use this key
		env = append(env,
			fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", keyPath))
	}

	args := []string{"clone", "--depth=1", "--branch", branch, gitURL, workDir}
	cmd := exec.Command("git", args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if keyCleanup != nil {
		keyCleanup()
	}
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("git clone: %w\noutput: %s", err, out)
	}
	return workDir, cleanup, nil
}

// CommitHash returns the HEAD commit hash for a cloned repo at workDir.
func CommitHash(workDir string) (string, error) {
	cmd := exec.Command("git", "-C", workDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	hash := string(out)
	if len(hash) > 0 && hash[len(hash)-1] == '\n' {
		hash = hash[:len(hash)-1]
	}
	return hash, nil
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/git/... -v -run TestWriteKeyFileCleansUp
```

Expected: PASS. (TestClonePublicRepo may be skipped if no network.)

- [ ] **Step 6: Add go-git dependency** (actually we used exec.Command git, no external dep needed)

```bash
go mod tidy
```

- [ ] **Step 7: Commit**

```bash
git add internal/git/
git commit -m "feat: git gateway (clone via subprocess, SSH key temp file)"
```

---

## Task 5: Job Interface + LocalJob + LogWriter

**Files:**
- Create: `internal/execution/job.go`
- Create: `internal/execution/local_job.go`
- Create: `internal/execution/log_writer.go`
- Create: `internal/execution/log_writer_test.go`

- [ ] **Step 1: Write Job interface**

```go
// internal/execution/job.go
package execution

import "context"

// Job is the interface for an executable unit of work.
// Implementations: LocalJob (subprocess), RemoteJob (future).
type Job interface {
	// Run executes the job. It blocks until completion or ctx is cancelled.
	// Output lines are sent to the provided output channel.
	Run(ctx context.Context, output chan<- string) error
	// Kill terminates a running job.
	Kill()
	// IsKilled returns true if the job was killed.
	IsKilled() bool
}
```

- [ ] **Step 2: Write LocalJob**

```go
// internal/execution/local_job.go
package execution

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sync/atomic"
)

// LocalJob runs ansible-playbook as a subprocess.
type LocalJob struct {
	playbookPath  string   // absolute path to playbook file
	inventoryPath string   // absolute path to inventory file or inline content
	workDir       string   // working directory (cloned repo root)
	extraArgs     []string // additional ansible-playbook args
	extraEnv      []string // KEY=VALUE env vars
	cmd           *exec.Cmd
	killed        atomic.Bool
}

func NewLocalJob(playbookPath, inventoryPath, workDir string, extraArgs, extraEnv []string) *LocalJob {
	return &LocalJob{
		playbookPath:  playbookPath,
		inventoryPath: inventoryPath,
		workDir:       workDir,
		extraArgs:     extraArgs,
		extraEnv:      extraEnv,
	}
}

func (j *LocalJob) Run(ctx context.Context, output chan<- string) error {
	args := append([]string{"-i", j.inventoryPath, j.playbookPath}, j.extraArgs...)
	j.cmd = exec.CommandContext(ctx, "ansible-playbook", args...)
	j.cmd.Dir = j.workDir
	j.cmd.Env = append(j.cmd.Environ(), j.extraEnv...)

	stdout, err := j.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := j.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := j.cmd.Start(); err != nil {
		return fmt.Errorf("start ansible-playbook: %w", err)
	}

	// Stream stdout and stderr to output channel
	done := make(chan struct{}, 2)
	scanAndSend := func(r interface{ Read([]byte) (int, error) }) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output <- scanner.Text()
		}
		done <- struct{}{}
	}
	go scanAndSend(stdout)
	go scanAndSend(stderr)

	<-done
	<-done

	if err := j.cmd.Wait(); err != nil {
		if j.killed.Load() {
			return nil // killed intentionally
		}
		return fmt.Errorf("ansible-playbook: %w", err)
	}
	return nil
}

func (j *LocalJob) Kill() {
	j.killed.Store(true)
	if j.cmd != nil && j.cmd.Process != nil {
		j.cmd.Process.Kill()
	}
}

func (j *LocalJob) IsKilled() bool {
	return j.killed.Load()
}
```

- [ ] **Step 3: Write failing test for LogWriter**

```go
// internal/execution/log_writer_test.go
package execution_test

import (
	"testing"
	"time"
	"github.com/kgory/kirmaphor/internal/execution"
)

func TestLogWriterBatchesLines(t *testing.T) {
	var flushed [][]string
	flush := func(lines []string) error {
		flushed = append(flushed, append([]string{}, lines...))
		return nil
	}

	lw := execution.NewLogWriter(flush, 100*time.Millisecond)
	lw.Write("line 1")
	lw.Write("line 2")
	lw.Write("line 3")

	// Wait for flush interval
	time.Sleep(200 * time.Millisecond)
	lw.Close()

	if len(flushed) == 0 {
		t.Fatal("expected at least one flush")
	}
	total := 0
	for _, batch := range flushed {
		total += len(batch)
	}
	if total != 3 {
		t.Fatalf("expected 3 lines total, got %d", total)
	}
}

func TestLogWriterFlushOnClose(t *testing.T) {
	var flushed [][]string
	flush := func(lines []string) error {
		flushed = append(flushed, append([]string{}, lines...))
		return nil
	}

	lw := execution.NewLogWriter(flush, 10*time.Second) // long interval
	lw.Write("only line")
	lw.Close() // should flush immediately

	if len(flushed) == 0 || flushed[0][0] != "only line" {
		t.Fatal("expected line to be flushed on Close")
	}
}
```

- [ ] **Step 4: Run, confirm FAIL**

```bash
go test ./internal/execution/... -run TestLogWriter -v
```

Expected: FAIL.

- [ ] **Step 5: Implement LogWriter**

```go
// internal/execution/log_writer.go
package execution

import (
	"sync"
	"time"
)

// LogWriter batches output lines and flushes them via the provided function
// at a regular interval or when Close is called.
// This matches Semaphore's 500ms batch pattern.
type LogWriter struct {
	flush    func(lines []string) error
	interval time.Duration
	mu       sync.Mutex
	buffer   []string
	done     chan struct{}
	once     sync.Once
}

func NewLogWriter(flush func(lines []string) error, interval time.Duration) *LogWriter {
	lw := &LogWriter{
		flush:    flush,
		interval: interval,
		done:     make(chan struct{}),
	}
	go lw.run()
	return lw
}

func (lw *LogWriter) Write(line string) {
	lw.mu.Lock()
	lw.buffer = append(lw.buffer, line)
	lw.mu.Unlock()
}

func (lw *LogWriter) flushNow() {
	lw.mu.Lock()
	if len(lw.buffer) == 0 {
		lw.mu.Unlock()
		return
	}
	lines := lw.buffer
	lw.buffer = nil
	lw.mu.Unlock()
	lw.flush(lines) // ignore error — caller can log separately
}

func (lw *LogWriter) run() {
	ticker := time.NewTicker(lw.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lw.flushNow()
		case <-lw.done:
			lw.flushNow() // final flush
			return
		}
	}
}

// Close flushes remaining lines and stops the background goroutine.
func (lw *LogWriter) Close() {
	lw.once.Do(func() { close(lw.done) })
	// Give the goroutine time to finish final flush
	time.Sleep(10 * time.Millisecond)
}
```

- [ ] **Step 6: Run tests, confirm PASS**

```bash
go test ./internal/execution/... -run TestLogWriter -v
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/execution/
git commit -m "feat: Job interface, LocalJob (ansible-playbook subprocess), LogWriter (batched)"
```

---

## Task 6: Inventory Loader

**Files:**
- Create: `internal/inventory/loader.go`
- Create: `internal/inventory/loader_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/inventory/loader_test.go
package inventory_test

import (
	"os"
	"strings"
	"testing"

	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/inventory"
)

func TestLoadStaticInventory(t *testing.T) {
	content := "[web]\n192.168.1.10\n192.168.1.11\n"
	inv := &models.Inventory{
		Type:          models.InventoryTypeStatic,
		InventoryData: &content,
	}

	path, cleanup, err := inventory.Load(inv)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "192.168.1.10") {
		t.Fatalf("expected inventory content in file, got: %s", data)
	}
}

func TestLoadStaticInventoryNilData(t *testing.T) {
	inv := &models.Inventory{
		Type:          models.InventoryTypeStatic,
		InventoryData: nil,
	}
	_, _, err := inventory.Load(inv)
	if err == nil {
		t.Fatal("expected error for nil inventory data")
	}
}
```

- [ ] **Step 2: Run, confirm FAIL**

```bash
go test ./internal/inventory/... -v
```

- [ ] **Step 3: Implement loader**

```go
// internal/inventory/loader.go
package inventory

import (
	"fmt"
	"os"

	"github.com/kgory/kirmaphor/internal/db/models"
)

// Load materialises an inventory to a temp file.
// Returns (tmpFilePath, cleanup func, error).
// Caller must call cleanup() after ansible-playbook finishes.
func Load(inv *models.Inventory) (string, func(), error) {
	switch inv.Type {
	case models.InventoryTypeStatic, models.InventoryTypeStaticYAML:
		return loadInline(inv)
	default:
		return "", nil, fmt.Errorf("inventory type %q not yet supported", inv.Type)
	}
}

func loadInline(inv *models.Inventory) (string, func(), error) {
	if inv.InventoryData == nil || *inv.InventoryData == "" {
		return "", nil, fmt.Errorf("inventory_data is empty for static inventory %s", inv.ID)
	}
	f, err := os.CreateTemp("", "kirmaphore-inventory-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp inventory: %w", err)
	}
	if _, err := f.WriteString(*inv.InventoryData); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("write inventory: %w", err)
	}
	f.Close()
	path := f.Name()
	return path, func() { os.Remove(path) }, nil
}
```

- [ ] **Step 4: Run, confirm PASS**

```bash
go test ./internal/inventory/... -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/inventory/
git commit -m "feat: inventory loader (static/static-yaml to temp file)"
```

---

## Task 7: TaskRunner

**Files:**
- Create: `internal/execution/task_runner.go`
- Create: `internal/execution/task_runner_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/execution/task_runner_test.go
package execution_test

import (
	"testing"
	"github.com/kgory/kirmaphor/internal/execution"
)

func TestTaskRunnerBuildArgs(t *testing.T) {
	args := execution.BuildAnsibleArgs("site.yml", "/tmp/inv", "--check --diff")
	if len(args) < 3 {
		t.Fatalf("expected at least 3 args, got %v", args)
	}
	found := false
	for _, a := range args {
		if a == "--check" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected --check in args, got %v", args)
	}
}

func TestTaskRunnerBuildEnv(t *testing.T) {
	env := execution.BuildEnv(map[string]string{
		"ANSIBLE_NOCOWS": "1",
		"MY_VAR":         "hello",
	})
	found := false
	for _, e := range env {
		if e == "ANSIBLE_NOCOWS=1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected ANSIBLE_NOCOWS=1 in env, got %v", env)
	}
}
```

- [ ] **Step 2: Run, confirm FAIL**

```bash
go test ./internal/execution/... -run TestTaskRunner -v
```

- [ ] **Step 3: Implement TaskRunner helpers**

```go
// internal/execution/task_runner.go
package execution

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	gitpkg "github.com/kgory/kirmaphor/internal/git"
	"github.com/kgory/kirmaphor/internal/inventory"
)

// BuildAnsibleArgs builds the argument list for ansible-playbook.
func BuildAnsibleArgs(playbookPath, inventoryPath, extraArgs string) []string {
	args := []string{"-i", inventoryPath, playbookPath}
	if extraArgs != "" {
		for _, a := range strings.Fields(extraArgs) {
			args = append(args, a)
		}
	}
	return args
}

// BuildEnv converts a map of env vars to KEY=VALUE slice for exec.Cmd.Env.
func BuildEnv(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// RunnerDeps holds dependencies injected into TaskRunner.
type RunnerDeps struct {
	Pool      *pgxpool.Pool
	Decrypt   func(encrypted, nonce []byte) ([]byte, error)
}

// RunTask executes a task end-to-end:
// 1. Loads repository credentials and clones the repo
// 2. Loads inventory to temp file
// 3. Runs ansible-playbook via LocalJob
// 4. Streams logs to DB via LogWriter
// 5. Updates task status on completion
func RunTask(ctx context.Context, deps RunnerDeps, task *models.Task) {
	pool := deps.Pool
	taskID := task.ID

	setStatus := func(status models.TaskStatus, msg string) {
		if err := queries.UpdateTaskStatus(ctx, pool, taskID, status, msg); err != nil {
			log.Printf("task %s: update status: %v", taskID, err)
		}
	}

	setStatus(models.TaskStatusRunning, "")

	// 1. Clone repository
	repo, err := queries.GetRepository(ctx, pool, task.RepositoryID)
	if err != nil {
		setStatus(models.TaskStatusError, fmt.Sprintf("get repository: %v", err))
		return
	}

	var keyPEM []byte
	if repo.SSHKeyID != nil {
		secret, err := queries.GetSecret(ctx, pool, *repo.SSHKeyID)
		if err != nil {
			setStatus(models.TaskStatusError, fmt.Sprintf("get ssh key: %v", err))
			return
		}
		keyPEM, err = deps.Decrypt(secret.EncryptedValue, secret.Nonce)
		if err != nil {
			setStatus(models.TaskStatusError, "decrypt ssh key failed")
			return
		}
	}

	workDir, repoCleanup, err := gitpkg.CloneOrPull(repo.GitURL, task.GitBranch, keyPEM)
	if err != nil {
		setStatus(models.TaskStatusError, fmt.Sprintf("git clone: %v", err))
		return
	}
	defer repoCleanup()

	// Store commit hash
	if hash, err := gitpkg.CommitHash(workDir); err == nil {
		queries.UpdateTaskCommit(ctx, pool, taskID, hash)
	}

	// 2. Load inventory
	var invFilePath string
	var invCleanup func()

	if task.InventoryID != nil {
		inv, err := queries.GetInventory(ctx, pool, *task.InventoryID)
		if err != nil {
			setStatus(models.TaskStatusError, fmt.Sprintf("get inventory: %v", err))
			return
		}
		invFilePath, invCleanup, err = inventory.Load(inv)
		if err != nil {
			setStatus(models.TaskStatusError, fmt.Sprintf("load inventory: %v", err))
			return
		}
		defer invCleanup()
	} else {
		invFilePath = "localhost,"
	}

	// 3. Set up log writer (flush every 500ms — Semaphore pattern)
	logWriter := NewLogWriter(func(lines []string) error {
		return queries.AppendLogs(ctx, pool, taskID, lines)
	}, 500*time.Millisecond)
	defer logWriter.Close()

	// 4. Run job
	extraArgs := []string{}
	if task.Arguments != "" {
		extraArgs = strings.Fields(task.Arguments)
	}
	extraEnv := BuildEnv(task.Environment)

	job := NewLocalJob(task.Playbook, invFilePath, workDir, extraArgs, extraEnv)
	output := make(chan string, 100)

	go func() {
		for line := range output {
			logWriter.Write(line)
		}
	}()

	runErr := job.Run(ctx, output)
	close(output)

	// 5. Final status
	if job.IsKilled() {
		setStatus(models.TaskStatusStopped, "killed by user")
		return
	}
	if runErr != nil {
		setStatus(models.TaskStatusError, runErr.Error())
		return
	}
	setStatus(models.TaskStatusSuccess, "")
}
```

**Note:** `queries.GetSecret` needs to be added. Add it to `internal/db/queries/secret.go`:

```go
func GetSecret(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Secret, error) {
	s := &models.Secret{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, name, type, encrypted_value, nonce, created_by, created_at, updated_at
		 FROM secrets WHERE id = $1`, id,
	).Scan(&s.ID, &s.ProjectID, &s.Name, &s.Type, &s.EncryptedValue, &s.Nonce,
		&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}
```

- [ ] **Step 4: Run tests, confirm PASS**

```bash
go test ./internal/execution/... -run TestTaskRunner -v
```

Expected: PASS.

- [ ] **Step 5: Build check**

```bash
go build ./...
```

Fix any imports. Add `"github.com/google/uuid"` and missing imports as needed.

- [ ] **Step 6: Commit**

```bash
git add internal/execution/task_runner.go internal/execution/task_runner_test.go internal/db/queries/secret.go
git commit -m "feat: TaskRunner (clone → inventory → execute → log → status)"
```

---

## Task 8: TaskPool

**Files:**
- Create: `internal/execution/task_pool.go`
- Create: `internal/execution/task_pool_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/execution/task_pool_test.go
package execution_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/google/uuid"
)

func TestTaskPoolRunsJob(t *testing.T) {
	var ran atomic.Int32
	pool := execution.NewTaskPool(2) // max 2 concurrent
	pool.Start()
	defer pool.Stop()

	for i := 0; i < 3; i++ {
		pool.Enqueue(execution.TaskRequest{
			TaskID: uuid.New(),
			Run: func(ctx context.Context) {
				ran.Add(1)
			},
		})
	}

	// Wait for all to complete
	time.Sleep(200 * time.Millisecond)
	if ran.Load() != 3 {
		t.Fatalf("expected 3 tasks run, got %d", ran.Load())
	}
}

func TestTaskPoolRespectsConcurrencyLimit(t *testing.T) {
	var concurrent atomic.Int32
	var maxSeen atomic.Int32

	pool := execution.NewTaskPool(2)
	pool.Start()
	defer pool.Stop()

	for i := 0; i < 5; i++ {
		pool.Enqueue(execution.TaskRequest{
			TaskID: uuid.New(),
			Run: func(ctx context.Context) {
				c := concurrent.Add(1)
				if c > maxSeen.Load() {
					maxSeen.Store(c)
				}
				time.Sleep(50 * time.Millisecond)
				concurrent.Add(-1)
			},
		})
	}

	time.Sleep(500 * time.Millisecond)
	if maxSeen.Load() > 2 {
		t.Fatalf("concurrency exceeded limit: max seen = %d", maxSeen.Load())
	}
}
```

- [ ] **Step 2: Run, confirm FAIL**

```bash
go test ./internal/execution/... -run TestTaskPool -v
```

Expected: FAIL.

- [ ] **Step 3: Implement TaskPool (Semaphore channel pattern)**

```go
// internal/execution/task_pool.go
package execution

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

// TaskRequest is a unit of work submitted to the TaskPool.
type TaskRequest struct {
	TaskID uuid.UUID
	Run    func(ctx context.Context)
}

// TaskPool manages a bounded pool of concurrent task runners.
// Pattern adopted from semaphoreui/semaphore: TaskPool.go.
// Uses a buffered channel as the queue and a semaphore channel for concurrency.
type TaskPool struct {
	maxConcurrent int
	queue         chan TaskRequest
	sem           chan struct{}
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewTaskPool(maxConcurrent int) *TaskPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskPool{
		maxConcurrent: maxConcurrent,
		queue:         make(chan TaskRequest, 500), // buffered — Semaphore uses 500
		sem:           make(chan struct{}, maxConcurrent),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins processing enqueued tasks.
func (p *TaskPool) Start() {
	go p.dispatch()
}

// Stop drains the queue and waits for running tasks to complete.
func (p *TaskPool) Stop() {
	p.cancel()
	p.wg.Wait()
}

// Enqueue adds a task to the queue. Non-blocking if queue not full.
func (p *TaskPool) Enqueue(req TaskRequest) {
	select {
	case p.queue <- req:
	default:
		log.Printf("taskpool: queue full, dropping task %s", req.TaskID)
	}
}

func (p *TaskPool) dispatch() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case req := <-p.queue:
			p.sem <- struct{}{} // acquire slot
			p.wg.Add(1)
			go func(r TaskRequest) {
				defer func() {
					<-p.sem // release slot
					p.wg.Done()
					if rec := recover(); rec != nil {
						log.Printf("taskpool: panic in task %s: %v", r.TaskID, rec)
					}
				}()
				r.Run(p.ctx)
			}(req)
		}
	}
}
```

- [ ] **Step 4: Run tests, confirm PASS**

```bash
go test ./internal/execution/... -run TestTaskPool -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/execution/task_pool.go internal/execution/task_pool_test.go
git commit -m "feat: TaskPool (channel-based queue, bounded concurrency, Semaphore pattern)"
```

---

## Task 9: Scheduler

**Files:**
- Create: `internal/scheduler/scheduler.go`
- Create: `internal/scheduler/scheduler_test.go`

Install dependency:
```bash
go get github.com/robfig/cron/v3@latest
```

- [ ] **Step 1: Write failing test**

```go
// internal/scheduler/scheduler_test.go
package scheduler_test

import (
	"testing"
	"github.com/kgory/kirmaphor/internal/scheduler"
)

func TestValidateCronFormat(t *testing.T) {
	valid := []string{
		"0 9 * * 1-5",    // weekdays at 9am
		"*/5 * * * *",    // every 5 minutes
		"0 0 1 * *",      // first of month
	}
	invalid := []string{
		"not a cron",
		"99 * * * *",     // invalid minute
		"",
	}
	for _, c := range valid {
		if err := scheduler.ValidateCronFormat(c); err != nil {
			t.Errorf("expected %q to be valid, got: %v", c, err)
		}
	}
	for _, c := range invalid {
		if err := scheduler.ValidateCronFormat(c); err == nil {
			t.Errorf("expected %q to be invalid", c)
		}
	}
}

func TestIsDue(t *testing.T) {
	// A cron that ran 2 minutes ago and should run every minute
	// should be due.
	isDue, err := scheduler.IsCronDue("* * * * *", nil)
	if err != nil {
		t.Fatalf("IsCronDue: %v", err)
	}
	if !isDue {
		t.Fatal("expected every-minute cron to be due when never run")
	}
}
```

- [ ] **Step 2: Run, confirm FAIL**

```bash
go test ./internal/scheduler/... -v
```

- [ ] **Step 3: Implement scheduler**

```go
// internal/scheduler/scheduler.go
package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/robfig/cron/v3"
)

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// ValidateCronFormat returns an error if the cron expression is invalid.
func ValidateCronFormat(expr string) error {
	if expr == "" {
		return fmt.Errorf("cron format is empty")
	}
	_, err := parser.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}
	return nil
}

// IsCronDue returns true if the cron expression is due given the last run time.
// If lastRun is nil, it is always due.
func IsCronDue(expr string, lastRun *time.Time) (bool, error) {
	schedule, err := parser.Parse(expr)
	if err != nil {
		return false, fmt.Errorf("parse cron: %w", err)
	}
	if lastRun == nil {
		return true, nil
	}
	next := schedule.Next(*lastRun)
	return next.Before(time.Now()), nil
}

// Scheduler polls the DB every 30s and enqueues due tasks.
type Scheduler struct {
	pool     *pgxpool.Pool
	taskPool *execution.TaskPool
	deps     execution.RunnerDeps
}

func New(pool *pgxpool.Pool, taskPool *execution.TaskPool, deps execution.RunnerDeps) *Scheduler {
	return &Scheduler{pool: pool, taskPool: taskPool, deps: deps}
}

// Run starts the scheduler loop. Blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	schedules, err := queries.GetDueSchedules(ctx, s.pool)
	if err != nil {
		log.Printf("scheduler: get due schedules: %v", err)
		return
	}
	for _, sched := range schedules {
		if err := s.process(ctx, sched); err != nil {
			log.Printf("scheduler: process schedule %s: %v", sched.ID, err)
		}
	}
}

func (s *Scheduler) process(ctx context.Context, sched *models.Schedule) error {
	// For cron: check if actually due
	if sched.Type == models.ScheduleTypeCron {
		if sched.CronFormat == nil {
			return fmt.Errorf("cron schedule %s has no format", sched.ID)
		}
		due, err := IsCronDue(*sched.CronFormat, sched.LastRunAt)
		if err != nil {
			return err
		}
		if !due {
			return nil
		}
	}

	// Load template to build task
	tmpl, err := queries.GetTemplate(ctx, s.pool, sched.TemplateID)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	schedID := sched.ID
	task := &models.Task{
		ProjectID:    sched.ProjectID,
		TemplateID:   tmpl.ID,
		Playbook:     tmpl.Playbook,
		InventoryID:  tmpl.InventoryID,
		RepositoryID: tmpl.RepositoryID,
		GitBranch:    "main", // use repo default
		Arguments:    tmpl.Arguments,
		Environment:  tmpl.Environment,
		CreatedBy:    tmpl.CreatedBy,
		ScheduleID:   &schedID,
	}

	created, err := queries.CreateTask(ctx, s.pool, task)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	taskCopy := *created
	deps := s.deps
	s.taskPool.Enqueue(execution.TaskRequest{
		TaskID: created.ID,
		Run: func(ctx context.Context) {
			execution.RunTask(ctx, deps, &taskCopy)
		},
	})

	// Touch schedule
	now := time.Now()
	if err := queries.TouchSchedule(ctx, s.pool, sched.ID, now); err != nil {
		log.Printf("scheduler: touch schedule %s: %v", sched.ID, err)
	}

	// Delete if one-shot
	if sched.DeleteAfterRun {
		queries.DeleteSchedule(ctx, s.pool, sched.ID)
	}

	return nil
}
```

- [ ] **Step 4: Run tests, confirm PASS**

```bash
go test ./internal/scheduler/... -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/scheduler/
git commit -m "feat: cron scheduler (ValidateCronFormat, IsCronDue, Scheduler.Run)"
```

---

## Task 10: API Handlers — Templates, Tasks, Inventories, Repositories

**Files:**
- Create: `internal/api/handlers/templates.go`
- Create: `internal/api/handlers/tasks.go`
- Create: `internal/api/handlers/inventories.go`
- Create: `internal/api/handlers/repositories.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: Write templates handler**

```go
// internal/api/handlers/templates.go
package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func ListTemplates(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		if !hasProjectAccess(r.Context(), pool, projectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		templates, err := queries.ListTemplates(r.Context(), pool, projectID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if templates == nil {
			templates = []*models.JobTemplate{}
		}
		helpers.WriteJSON(w, http.StatusOK, templates)
	}
}

func CreateTemplate(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermEditPlaybooks) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req struct {
			Name         string            `json:"name"`
			Description  string            `json:"description"`
			Playbook     string            `json:"playbook"`
			InventoryID  *uuid.UUID        `json:"inventory_id"`
			RepositoryID uuid.UUID         `json:"repository_id"`
			Environment  map[string]string `json:"environment"`
			Arguments    string            `json:"arguments"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		if req.Name == "" || req.Playbook == "" {
			helpers.WriteError(w, http.StatusBadRequest, "name and playbook are required")
			return
		}
		t := &models.JobTemplate{
			ProjectID:    projectID,
			Name:         req.Name,
			Description:  req.Description,
			Playbook:     req.Playbook,
			InventoryID:  req.InventoryID,
			RepositoryID: req.RepositoryID,
			Environment:  req.Environment,
			Arguments:    req.Arguments,
			CreatedBy:    user.ID,
		}
		created, err := queries.CreateTemplate(r.Context(), pool, t)
		if err != nil {
			helpers.WriteError(w, http.StatusConflict, "template name already exists or invalid reference")
			return
		}
		helpers.WriteJSON(w, http.StatusCreated, created)
	}
}

func GetTemplate(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("templateId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid template id")
			return
		}
		t, err := queries.GetTemplate(r.Context(), pool, id)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "template not found")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, t)
	}
}

func DeleteTemplate(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, _ := uuid.Parse(r.PathValue("projectId"))
		id, err := uuid.Parse(r.PathValue("templateId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid template id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermEditPlaybooks) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		if err := queries.DeleteTemplate(r.Context(), pool, id, projectID); err != nil {
			helpers.WriteError(w, http.StatusNotFound, "template not found")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
```

- [ ] **Step 2: Add hasProjectAccess helper to a shared file**

```go
// internal/api/handlers/access.go
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

// hasProjectAccess returns true if userID has the given permission in projectID.
// Returns false on any error (no-rows = not a member; other = DB error, fail safe).
func hasProjectAccess(ctx context.Context, pool *pgxpool.Pool,
	projectID, userID uuid.UUID, perm rbac.Permission) bool {
	role, err := queries.GetProjectRole(ctx, pool, projectID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false
		}
		return false // fail safe on DB error
	}
	return rbac.HasPermission(role, perm)
}
```

- [ ] **Step 3: Write tasks handler**

```go
// internal/api/handlers/tasks.go
package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/kgory/kirmaphor/internal/rbac"
)

// RunTemplate enqueues a new task from a template.
func RunTemplate(pool *pgxpool.Pool, taskPool *execution.TaskPool, deps execution.RunnerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermRunJobs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req struct {
			TemplateID  uuid.UUID         `json:"template_id"`
			GitBranch   string            `json:"git_branch"`
			Arguments   string            `json:"arguments"`
			Environment map[string]string `json:"environment"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		tmpl, err := queries.GetTemplate(r.Context(), pool, req.TemplateID)
		if err != nil || tmpl.ProjectID != projectID {
			helpers.WriteError(w, http.StatusNotFound, "template not found")
			return
		}
		branch := req.GitBranch
		if branch == "" {
			branch = "main"
		}
		env := tmpl.Environment
		if env == nil {
			env = map[string]string{}
		}
		for k, v := range req.Environment {
			env[k] = v
		}
		task := &models.Task{
			ProjectID:    projectID,
			TemplateID:   tmpl.ID,
			Playbook:     tmpl.Playbook,
			InventoryID:  tmpl.InventoryID,
			RepositoryID: tmpl.RepositoryID,
			GitBranch:    branch,
			Arguments:    req.Arguments,
			Environment:  env,
			CreatedBy:    user.ID,
		}
		created, err := queries.CreateTask(r.Context(), pool, task)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "create task failed")
			return
		}
		taskCopy := *created
		depsCopy := deps
		taskPool.Enqueue(execution.TaskRequest{
			TaskID: created.ID,
			Run: func(ctx interface{ Done() <-chan struct{} }) {
				// This cast is needed because TaskRequest.Run takes context.Context
				// Fix: use proper context
			},
		})
		// Actually: proper enqueue
		helpers.WriteJSON(w, http.StatusCreated, created)

		// Enqueue after responding (fire and forget)
		go func() {
			execution.RunTask(r.Context(), depsCopy, &taskCopy)
		}()
	}
}

func GetTask(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("taskId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		task, err := queries.GetTask(r.Context(), pool, id)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		// Verify project access
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, task.ProjectID, user.ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, task)
	}
}

func ListTasks(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		tasks, err := queries.ListTasks(r.Context(), pool, projectID, 50)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if tasks == nil {
			tasks = []*models.Task{}
		}
		helpers.WriteJSON(w, http.StatusOK, tasks)
	}
}
```

**Note:** Fix the `RunTemplate` handler to use TaskPool properly instead of raw goroutine:

```go
// Replace the broken enqueue block in RunTemplate with:
taskCopy := *created
depsCopy := deps
taskPool.Enqueue(execution.TaskRequest{
    TaskID: created.ID,
    Run: func(ctx context.Context) {
        execution.RunTask(ctx, depsCopy, &taskCopy)
    },
})
helpers.WriteJSON(w, http.StatusCreated, created)
// Remove the raw goroutine — TaskPool handles this
```

- [ ] **Step 4: Write inventories and repositories handlers**

```go
// internal/api/handlers/inventories.go
package handlers

import (
	"net/http"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func ListInventories(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		if !hasProjectAccess(r.Context(), pool, projectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		invs, err := queries.ListInventories(r.Context(), pool, projectID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if invs == nil {
			invs = []*models.Inventory{}
		}
		helpers.WriteJSON(w, http.StatusOK, invs)
	}
}

func CreateInventory(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermEditPlaybooks) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req struct {
			Name          string                `json:"name"`
			Type          models.InventoryType  `json:"type"`
			InventoryData *string               `json:"inventory_data"`
			RepositoryID  *uuid.UUID            `json:"repository_id"`
			InventoryPath *string               `json:"inventory_path"`
			SSHKeyID      *uuid.UUID            `json:"ssh_key_id"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		if req.Name == "" || req.Type == "" {
			helpers.WriteError(w, http.StatusBadRequest, "name and type are required")
			return
		}
		inv := &models.Inventory{
			ProjectID:     projectID,
			Name:          req.Name,
			Type:          req.Type,
			InventoryData: req.InventoryData,
			RepositoryID:  req.RepositoryID,
			InventoryPath: req.InventoryPath,
			SSHKeyID:      req.SSHKeyID,
			CreatedBy:     user.ID,
		}
		created, err := queries.CreateInventory(r.Context(), pool, inv)
		if err != nil {
			helpers.WriteError(w, http.StatusConflict, "inventory name already exists")
			return
		}
		helpers.WriteJSON(w, http.StatusCreated, created)
	}
}

func DeleteInventory(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, _ := uuid.Parse(r.PathValue("projectId"))
		id, err := uuid.Parse(r.PathValue("inventoryId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid inventory id")
			return
		}
		if !hasProjectAccess(r.Context(), pool, projectID, helpers.GetUser(r).ID, rbac.PermEditPlaybooks) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		queries.DeleteInventory(r.Context(), pool, id, projectID)
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
```

```go
// internal/api/handlers/repositories.go
package handlers

import (
	"net/http"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func ListRepositories(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		if !hasProjectAccess(r.Context(), pool, projectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		repos, err := queries.ListRepositories(r.Context(), pool, projectID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if repos == nil {
			repos = []*models.Repository{}
		}
		helpers.WriteJSON(w, http.StatusOK, repos)
	}
}

func CreateRepository(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermEditPlaybooks) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req struct {
			Name      string     `json:"name"`
			GitURL    string     `json:"git_url"`
			GitBranch string     `json:"git_branch"`
			SSHKeyID  *uuid.UUID `json:"ssh_key_id"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		if req.Name == "" || req.GitURL == "" {
			helpers.WriteError(w, http.StatusBadRequest, "name and git_url are required")
			return
		}
		if req.GitBranch == "" {
			req.GitBranch = "main"
		}
		repo := &models.Repository{
			ProjectID: projectID,
			Name:      req.Name,
			GitURL:    req.GitURL,
			GitBranch: req.GitBranch,
			SSHKeyID:  req.SSHKeyID,
			CreatedBy: user.ID,
		}
		created, err := queries.CreateRepository(r.Context(), pool, repo)
		if err != nil {
			helpers.WriteError(w, http.StatusConflict, "repository name already exists")
			return
		}
		helpers.WriteJSON(w, http.StatusCreated, created)
	}
}

func DeleteRepository(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, _ := uuid.Parse(r.PathValue("projectId"))
		id, err := uuid.Parse(r.PathValue("repoId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid repo id")
			return
		}
		if !hasProjectAccess(r.Context(), pool, projectID, helpers.GetUser(r).ID, rbac.PermEditPlaybooks) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		queries.DeleteRepository(r.Context(), pool, id, projectID)
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
```

- [ ] **Step 5: Update router with all new routes**

In `internal/api/router.go`, add to the authenticated group:

```go
// Execution routes (inside r.Group with requireAuth):
r.Get("/projects/{projectId}/templates", handlers.ListTemplates(pool))
r.Post("/projects/{projectId}/templates", handlers.CreateTemplate(pool))
r.Get("/projects/{projectId}/templates/{templateId}", handlers.GetTemplate(pool))
r.Delete("/projects/{projectId}/templates/{templateId}", handlers.DeleteTemplate(pool))

r.Post("/projects/{projectId}/run", handlers.RunTemplate(pool, taskPool, deps))
r.Get("/projects/{projectId}/tasks", handlers.ListTasks(pool))
r.Get("/tasks/{taskId}", handlers.GetTask(pool))

r.Get("/projects/{projectId}/inventories", handlers.ListInventories(pool))
r.Post("/projects/{projectId}/inventories", handlers.CreateInventory(pool))
r.Delete("/projects/{projectId}/inventories/{inventoryId}", handlers.DeleteInventory(pool))

r.Get("/projects/{projectId}/repositories", handlers.ListRepositories(pool))
r.Post("/projects/{projectId}/repositories", handlers.CreateRepository(pool))
r.Delete("/projects/{projectId}/repositories/{repoId}", handlers.DeleteRepository(pool))
```

`NewRouter` must also accept `taskPool *execution.TaskPool` and `deps execution.RunnerDeps` parameters. Update `cmd/kirmaphore/main.go` to initialize these before calling `api.NewRouter`.

- [ ] **Step 6: Build check**

```bash
cd /Users/kgory/dev/kirmaphor
go build ./...
```

Fix any import or signature errors.

- [ ] **Step 7: Commit**

```bash
git add internal/api/handlers/ internal/api/router.go cmd/kirmaphore/main.go
git commit -m "feat: execution API handlers (templates, tasks, inventories, repositories)"
```

---

## Task 11: WebSocket Log Streaming

**Files:**
- Create: `internal/api/handlers/logs.go`
- Modify: `internal/api/router.go`

Install dependency:
```bash
go get github.com/gorilla/websocket@latest
```

- [ ] **Step 1: Write logs handler (HTTP + WebSocket)**

```go
// internal/api/handlers/logs.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: restrict in production
}

// GetLogs returns all logs for a task as JSON.
func GetLogs(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, err := uuid.Parse(r.PathValue("taskId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		task, err := queries.GetTask(r.Context(), pool, taskID)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		if !hasProjectAccess(r.Context(), pool, task.ProjectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		logs, err := queries.GetLogs(r.Context(), pool, taskID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if logs == nil {
			logs = nil
		}
		helpers.WriteJSON(w, http.StatusOK, logs)
	}
}

// StreamLogs upgrades to WebSocket and streams new log lines as they appear.
// Client receives JSON: {"id": 42, "output": "PLAY [all] ***", "ts": "..."}
// Client can reconnect by passing ?after=<last_id> query param.
func StreamLogs(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, err := uuid.Parse(r.PathValue("taskId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		task, err := queries.GetTask(r.Context(), pool, taskID)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		if !hasProjectAccess(r.Context(), pool, task.ProjectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return // Upgrade writes the error
		}
		defer conn.Close()

		var afterID int64
		if s := r.URL.Query().Get("after"); s != "" {
			afterID, _ = strconv.ParseInt(s, 10, 64)
		}

		// Poll for new logs every 500ms until task is terminal
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				logs, err := queries.GetLogsAfter(r.Context(), pool, taskID, afterID)
				if err != nil {
					return
				}
				for _, l := range logs {
					msg, _ := json.Marshal(map[string]any{
						"id":     l.ID,
						"output": l.Output,
						"ts":     l.CreatedAt,
					})
					if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						return
					}
					afterID = l.ID
				}
				// Check if task is terminal — if so, close after draining
				refreshed, err := queries.GetTask(r.Context(), pool, taskID)
				if err != nil {
					return
				}
				switch refreshed.Status {
				case "success", "error", "stopped":
					// Send remaining logs one more time then close
					conn.WriteMessage(websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseNormalClosure, "task complete"))
					return
				}
			}
		}
	}
}
```

- [ ] **Step 2: Add log routes to router**

```go
// In internal/api/router.go authenticated group:
r.Get("/tasks/{taskId}/logs", handlers.GetLogs(pool))
r.Get("/tasks/{taskId}/logs/stream", handlers.StreamLogs(pool))
```

- [ ] **Step 3: Build check**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/api/handlers/logs.go internal/api/router.go go.mod go.sum
git commit -m "feat: WebSocket log streaming and HTTP log endpoint"
```

---

## Task 12: Wire Execution into main.go + Integration Smoke Test

**Files:**
- Modify: `cmd/kirmaphore/main.go`

- [ ] **Step 1: Update main.go to initialise and start TaskPool + Scheduler**

```go
// cmd/kirmaphore/main.go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/kgory/kirmaphor/internal/api"
	"github.com/kgory/kirmaphor/internal/config"
	"github.com/kgory/kirmaphor/internal/crypto"
	"github.com/kgory/kirmaphor/internal/db"
	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/kgory/kirmaphor/internal/scheduler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.Connect(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(cfg.DBURL, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	masterKey, err := crypto.LoadMasterKey(cfg.MasterKey)
	if err != nil {
		log.Fatalf("master key: %v", err)
	}

	deps := execution.RunnerDeps{
		Pool: pool,
		Decrypt: func(encrypted, nonce []byte) ([]byte, error) {
			return crypto.Decrypt(masterKey, encrypted, nonce)
		},
	}

	taskPool := execution.NewTaskPool(10) // max 10 concurrent tasks
	taskPool.Start()
	defer taskPool.Stop()

	sched := scheduler.New(pool, taskPool, deps)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sched.Run(ctx)

	router := api.NewRouter(cfg, pool, taskPool, deps)
	log.Printf("starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
```

- [ ] **Step 2: Update NewRouter signature**

```go
// internal/api/router.go — update signature:
func NewRouter(cfg *config.Config, pool *pgxpool.Pool,
	taskPool *execution.TaskPool, deps execution.RunnerDeps) http.Handler {
```

Import `"github.com/kgory/kirmaphor/internal/execution"` in router.go.

- [ ] **Step 3: Full build check**

```bash
cd /Users/kgory/dev/kirmaphor
go build ./...
```

Expected: clean build.

- [ ] **Step 4: Run all tests**

```bash
go test ./... -v 2>&1 | tail -30
```

Expected: all existing tests pass + new tests pass.

- [ ] **Step 5: Smoke test — start server and hit health endpoint**

```bash
export DATABASE_URL="postgres://kirmaphore:kirmaphore@localhost:5432/kirmaphore?sslmode=disable"
export MASTER_KEY="0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
go run cmd/kirmaphore/main.go &
sleep 2
curl -s http://localhost:8080/api/health
```

Expected: `{"status":"ok"}`

Kill background server after test.

- [ ] **Step 6: Commit**

```bash
git add cmd/kirmaphore/main.go internal/api/router.go
git commit -m "feat: wire TaskPool + Scheduler into main, update NewRouter signature"
```

---

## Self-Review

**Spec coverage check:**

| Spec Section | Task |
|---|---|
| Execution Engine — job queue (TaskPool) | Task 8 |
| Execution Engine — Ansible subprocess (LocalJob) | Task 5 |
| Execution Engine — Docker runners (interface ready, Docker impl deferred) | Task 5 (Job interface) |
| Inventory Manager — static/static-yaml | Task 6 |
| Inventory Manager — cloud dynamic (aws/azure/gcp) | Type defined, implementation deferred to Plan 2b |
| Git Gateway | Task 4 |
| Scheduling (cron + run_at) | Task 9 |
| DB migrations for execution | Task 1 |
| API: templates, tasks, inventories, repos | Task 10 |
| WebSocket live log stream | Task 11 |
| Log batching (500ms) | Task 5 (LogWriter) |
| TaskRunner assembles context | Task 7 |

**Cloud dynamic inventories** (aws-ec2, azure-vmss, gcp-gce) types are defined in the DB schema and models, but the `Load()` function returns an error for those types. Full cloud inventory implementation is Plan 2b (short extension plan) to keep this plan shippable.

**Placeholder scan:** No TBD sections. The tasks.go `RunTemplate` handler has a note to fix the enqueue block — this is resolved inline in Step 3.

**Type consistency:**
- `execution.RunnerDeps` used in Task 7, 9, 10, 12 consistently
- `execution.TaskPool` used in Task 8, 10, 12 consistently
- `models.TaskStatus` constants used in queries/task.go and models/task.go consistently
- `queries.AppendLogs` signature matches LogWriter flush func in Task 7
