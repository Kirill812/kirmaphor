package handlers

import (
	"errors"
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
			Name          string               `json:"name"`
			Type          models.InventoryType `json:"type"`
			InventoryData *string              `json:"inventory_data"`
			RepositoryID  *uuid.UUID           `json:"repository_id"`
			InventoryPath *string              `json:"inventory_path"`
			SSHKeyID      *uuid.UUID           `json:"ssh_key_id"`
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
		if err := queries.DeleteInventory(r.Context(), pool, id, projectID); err != nil {
			if errors.Is(err, queries.ErrNotFound) {
				helpers.WriteError(w, http.StatusNotFound, "inventory not found")
			} else {
				helpers.WriteError(w, http.StatusInternalServerError, "server error")
			}
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
