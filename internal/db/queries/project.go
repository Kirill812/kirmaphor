package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func CreateProject(ctx context.Context, pool *pgxpool.Pool, name, description string, createdBy uuid.UUID) (*models.Project, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	p := &models.Project{}
	err = tx.QueryRow(ctx,
		`INSERT INTO projects (name, description, created_by) VALUES ($1, $2, $3)
		 RETURNING id, name, description, created_by, created_at`,
		name, description, createdBy,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO project_users (project_id, user_id, role) VALUES ($1, $2, $3)`,
		p.ID, createdBy, rbac.RoleOwner)
	if err != nil {
		return nil, err
	}
	return p, tx.Commit(ctx)
}

func GetProjectsByUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]*models.Project, error) {
	rows, err := pool.Query(ctx,
		`SELECT p.id, p.name, p.description, p.created_by, p.created_at
		 FROM projects p
		 JOIN project_users pu ON pu.project_id = p.id
		 WHERE pu.user_id = $1
		 ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func GetProjectByID(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) (*models.Project, error) {
	p := &models.Project{}
	err := pool.QueryRow(ctx,
		`SELECT id, name, description, created_by, created_at FROM projects WHERE id = $1`, projectID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func GetProjectRole(ctx context.Context, pool *pgxpool.Pool, projectID, userID uuid.UUID) (rbac.Role, error) {
	var role rbac.Role
	err := pool.QueryRow(ctx,
		`SELECT role FROM project_users WHERE project_id = $1 AND user_id = $2`,
		projectID, userID,
	).Scan(&role)
	return role, err
}
