package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/rbac"
)

// hasProjectAccess returns true if userID has the given permission in projectID.
// Returns false on any error (fail-safe).
func hasProjectAccess(ctx context.Context, pool *pgxpool.Pool,
	projectID, userID uuid.UUID, perm rbac.Permission) bool {
	role, err := queries.GetProjectRole(ctx, pool, projectID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false
		}
		return false // fail safe on DB error
	}
	return rbac.HasPermission(role, perm)
}
