package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/kgory/kirmaphor/internal/db/models"
)

type contextKey string

const (
	CtxUser    contextKey = "user"
	CtxSession contextKey = "session"
	CtxProject contextKey = "project"
	CtxRole    contextKey = "role"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func Bind(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func GetUser(r *http.Request) *models.User {
	u, _ := r.Context().Value(CtxUser).(*models.User)
	return u
}

func GetSession(r *http.Request) *models.UserSession {
	s, _ := r.Context().Value(CtxSession).(*models.UserSession)
	return s
}
