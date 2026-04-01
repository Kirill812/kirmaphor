package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
)

func ListProjects(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		projects, err := queries.GetProjectsByUser(r.Context(), pool, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if projects == nil {
			projects = []*models.Project{}
		}
		helpers.WriteJSON(w, http.StatusOK, projects)
	}
}

func CreateProject(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		if req.Name == "" {
			helpers.WriteError(w, http.StatusBadRequest, "name is required")
			return
		}
		p, err := queries.CreateProject(r.Context(), pool, req.Name, req.Description, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusCreated, p)
	}
}

func GetProject(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		role, err := queries.GetProjectRole(r.Context(), pool, id, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "project not found or access denied")
			return
		}
		p, err := queries.GetProjectByID(r.Context(), pool, id)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "project not found")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]any{"project": p, "role": role})
	}
}
