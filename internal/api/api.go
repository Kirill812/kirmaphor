package api

import (
	"net/http"

	"github.com/kgory/kirmaphor/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter is a stub — to be implemented in Task 8.
func NewRouter(_ *config.Config, _ *pgxpool.Pool) http.Handler {
	return http.NewServeMux()
}
