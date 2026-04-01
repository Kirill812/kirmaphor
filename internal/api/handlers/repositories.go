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
		if err := queries.DeleteRepository(r.Context(), pool, id, projectID); err != nil {
			helpers.WriteError(w, http.StatusNotFound, "repository not found")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
