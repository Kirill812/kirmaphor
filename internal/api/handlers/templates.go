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
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		id, err := uuid.Parse(r.PathValue("templateId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid template id")
			return
		}
		if !hasProjectAccess(r.Context(), pool, projectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		t, err := queries.GetTemplate(r.Context(), pool, id)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "template not found")
			return
		}
		if t.ProjectID != projectID {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
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
			if errors.Is(err, queries.ErrNotFound) {
				helpers.WriteError(w, http.StatusNotFound, "template not found")
			} else {
				helpers.WriteError(w, http.StatusInternalServerError, "server error")
			}
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
