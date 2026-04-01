package queries

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
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
		return ErrNotFound
	}
	return nil
}
