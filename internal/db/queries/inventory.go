package queries

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
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
	tag, err := pool.Exec(ctx, `DELETE FROM inventories WHERE id = $1 AND project_id = $2`, id, projectID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
