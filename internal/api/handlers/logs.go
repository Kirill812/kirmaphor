package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: restrict in production
}

// GetLogs returns all logs for a task as JSON.
func GetLogs(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, err := uuid.Parse(r.PathValue("taskId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		task, err := queries.GetTask(r.Context(), pool, taskID)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		if !hasProjectAccess(r.Context(), pool, task.ProjectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		logs, err := queries.GetLogs(r.Context(), pool, taskID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if logs == nil {
			logs = []*models.TaskLog{}
		}
		helpers.WriteJSON(w, http.StatusOK, logs)
	}
}

// StreamLogs upgrades to WebSocket and streams new log lines as they appear.
// Client receives JSON: {"id": 42, "output": "PLAY [all] ***", "ts": "..."}
// Client can reconnect by passing ?after=<last_id> query param.
func StreamLogs(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, err := uuid.Parse(r.PathValue("taskId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		task, err := queries.GetTask(r.Context(), pool, taskID)
		if err != nil {
			helpers.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		if !hasProjectAccess(r.Context(), pool, task.ProjectID, helpers.GetUser(r).ID, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return // Upgrade writes the error
		}
		defer conn.Close()

		var afterID int64
		if s := r.URL.Query().Get("after"); s != "" {
			afterID, _ = strconv.ParseInt(s, 10, 64)
		}

		// Poll for new logs every 500ms until task is terminal
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				logs, err := queries.GetLogsAfter(r.Context(), pool, taskID, afterID)
				if err != nil {
					return
				}
				for _, l := range logs {
					msg, _ := json.Marshal(map[string]any{
						"id":     l.ID,
						"output": l.Output,
						"ts":     l.CreatedAt,
					})
					if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						return
					}
					afterID = l.ID
				}
				// Check if task is terminal — close after draining
				refreshed, err := queries.GetTask(r.Context(), pool, taskID)
				if err != nil {
					return
				}
				switch refreshed.Status {
				case models.TaskStatusSuccess, models.TaskStatusError, models.TaskStatusStopped:
					conn.WriteMessage(websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseNormalClosure, "task complete"))
					return
				}
			}
		}
	}
}
