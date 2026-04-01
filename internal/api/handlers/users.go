package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/db/queries"
)

func GetMe(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		helpers.WriteJSON(w, http.StatusOK, map[string]any{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"avatar_url":   user.AvatarURL,
			"onboarded":    user.Onboarded,
		})
	}
}

func UpdateMe(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		var req struct {
			DisplayName *string `json:"display_name"`
			AvatarURL   *string `json:"avatar_url"`
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		_, err := pool.Exec(r.Context(),
			`UPDATE users SET display_name = COALESCE($1, display_name),
			                  avatar_url = COALESCE($2, avatar_url)
			 WHERE id = $3`,
			req.DisplayName, req.AvatarURL, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func ListSessions(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		rows, err := pool.Query(r.Context(),
			`SELECT id, device_label, ip_address, is_current, expires_at, created_at
			 FROM user_sessions WHERE user_id = $1 AND is_current = TRUE
			 ORDER BY created_at DESC`, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		defer rows.Close()
		var sessions []map[string]any
		for rows.Next() {
			var id, deviceLabel, ipAddress any
			var isCurrent bool
			var expiresAt, createdAt any
			rows.Scan(&id, &deviceLabel, &ipAddress, &isCurrent, &expiresAt, &createdAt)
			sessions = append(sessions, map[string]any{
				"id":           id,
				"device_label": deviceLabel,
				"ip":           ipAddress,
				"is_current":   isCurrent,
				"expires_at":   expiresAt,
			})
		}
		if sessions == nil {
			sessions = []map[string]any{}
		}
		helpers.WriteJSON(w, http.StatusOK, sessions)
	}
}

func RevokeUserSession(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetUser(r)
		sessionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid session id")
			return
		}
		if err := queries.RevokeSession(r.Context(), pool, sessionID, user.ID); err != nil {
			if err == queries.ErrSessionNotFound {
				helpers.WriteError(w, http.StatusNotFound, "session not found")
				return
			}
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
	}
}
