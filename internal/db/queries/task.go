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
