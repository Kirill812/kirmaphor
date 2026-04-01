package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	"github.com/kgory/kirmaphor/internal/auth"
	"github.com/kgory/kirmaphor/internal/crypto"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func CreateSecret(pool *pgxpool.Pool, masterKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := helpers.GetSession(r)
		if err := auth.CheckSecureSession(session); err != nil {
			helpers.WriteError(w, http.StatusForbidden, "secure_session_required")
			return
		}
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		role, err := queries.GetProjectRole(r.Context(), pool, projectID, user.ID)
		if err != nil || !rbac.HasPermission(role, rbac.PermManageSecrets) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value []byte `json:"value"` // already Layer-1 client-encrypted
		}
		if !helpers.Bind(w, r, &req) {
			return
		}
		if req.Name == "" || req.Type == "" {
			helpers.WriteError(w, http.StatusBadRequest, "name and type are required")
			return
		}
		// Layer 2: server-side encrypt
		encrypted, nonce, err := crypto.Encrypt(masterKey, req.Value)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "encryption error")
			return
		}
		secret, err := queries.CreateSecret(r.Context(), pool, projectID, req.Name, req.Type, encrypted, nonce, user.ID)
		if err != nil {
			helpers.WriteError(w, http.StatusConflict, "secret already exists or invalid type")
			return
		}
		helpers.WriteJSON(w, http.StatusCreated, map[string]any{
			"id": secret.ID, "name": secret.Name, "type": secret.Type,
		})
	}
}

func ListSecrets(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(r.PathValue("projectId"))
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		user := helpers.GetUser(r)
		role, err := queries.GetProjectRole(r.Context(), pool, projectID, user.ID)
		if err != nil || !rbac.HasPermission(role, rbac.PermReadLogs) {
			helpers.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		secrets, err := queries.ListSecrets(r.Context(), pool, projectID)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "server error")
			return
		}
		if secrets == nil {
			secrets = []*models.Secret{}
		}
		helpers.WriteJSON(w, http.StatusOK, secrets)
	}
}
