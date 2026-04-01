package queries

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
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
	tag, err := pool.Exec(ctx, `DELETE FROM repositories WHERE id = $1 AND project_id = $2`, id, projectID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
