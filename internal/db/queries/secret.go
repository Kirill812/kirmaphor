package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateSecret(ctx context.Context, pool *pgxpool.Pool,
	projectID uuid.UUID, name, secretType string,
	encryptedValue, nonce []byte, createdBy uuid.UUID,
) (*models.Secret, error) {
	s := &models.Secret{}
	err := pool.QueryRow(ctx,
		`INSERT INTO secrets (project_id, name, type, encrypted_value, nonce, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, project_id, name, type, encrypted_value, nonce, created_by, created_at, updated_at`,
		projectID, name, secretType, encryptedValue, nonce, createdBy,
	).Scan(&s.ID, &s.ProjectID, &s.Name, &s.Type, &s.EncryptedValue, &s.Nonce, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func ListSecrets(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) ([]*models.Secret, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, name, type, created_by, created_at, updated_at
		 FROM secrets WHERE project_id = $1 ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var secrets []*models.Secret
	for rows.Next() {
		s := &models.Secret{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &s.Type, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
	}
	return secrets, rows.Err()
}
