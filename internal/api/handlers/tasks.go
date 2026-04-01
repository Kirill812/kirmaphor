package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func RunTemplate(pool *pgxpool.Pool, taskPool *execution.TaskPool, deps execution.RunnerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermRunJobs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req struct {
			TemplateID  uuid.UUID         `json:"template_id"`
			GitBranch   string            `json:"git_branch"`
			Arguments   string            `json:"arguments"`
			Environment map[string]string `json:"environment"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		tmpl, err := queries.GetTemplate(r.Context(), pool, req.TemplateID)
		if err != nil || tmpl.ProjectID != projectID {
			helpers.WriteError(w, http.StatusNotFound, "template not found")
			return
		}
		branch := req.GitBranch
		if branch == "" {
			branch = "main"
		}
		env := tmpl.Environment
		if env == nil {
			env = map[string]string{}
		}
		for k, v := range req.Environment {
			env[k] = v
		}
		task := &models.Task{
			ProjectID:    projectID,
			TemplateID:   tmpl.ID,
			Playbook:     tmpl.Playbook,
			InventoryID:  tmpl.InventoryID,
			RepositoryID: tmpl.RepositoryID,
			GitBranch:    branch,
			Arguments:    req.Arguments,
			Environment:  env,
			CreatedBy:    user.ID,
		}
		created, err := queries.CreateTask(r.Context(), pool, task)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "create task failed")
			return
		}
		taskCopy := *created
		depsCopy := deps
		taskPool.Enqueue(execution.TaskRequest{
			TaskID: created.ID,
			Run: func(ctx context.Context) {
				execution.RunTask(ctx, depsCopy, &taskCopy)
			},
		})
		helpers.WriteJSON(w, http.StatusCreated, created)
	}
}

func GetTask(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("taskId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		task, err := queries.GetTask(r.Context(), pool, id)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, task.ProjectID, user.ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, task)
	}
}

func ListTasks(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		if !hasProjectAccess(r.Context(), pool, projectID, user.ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		tasks, err := queries.ListTasks(r.Context(), pool, projectID, 50)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if tasks == nil {
			tasks = []*models.Task{}
		}
		helpers.WriteJSON(w, http.StatusOK, tasks)
	}
}
