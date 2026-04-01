package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/api/helpers"
	apiMiddleware "github.com/kgory/kirmaphor/internal/api/middleware"
	"github.com/kgory/kirmaphor/internal/config"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)

	requireAuth := apiMiddleware.RequireAuth(pool)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		// Auth routes (public) — handlers added in Task 9
		r.Post("/auth/register", http.NotFound)
		r.Post("/auth/login", http.NotFound)
		r.Post("/auth/passkey/register/begin", http.NotFound)
		r.Post("/auth/passkey/register/finish", http.NotFound)
		r.Post("/auth/passkey/login/begin", http.NotFound)
		r.Post("/auth/passkey/login/finish", http.NotFound)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(requireAuth)
			r.Post("/auth/logout", http.NotFound)
			r.Get("/users/me", http.NotFound)
			r.Put("/users/me", http.NotFound)
			r.Get("/users/me/sessions", http.NotFound)
			r.Delete("/users/me/sessions/{id}", http.NotFound)
			r.Get("/projects", http.NotFound)
			r.Post("/projects", http.NotFound)
			r.Get("/projects/{id}", http.NotFound)
		})
	})

	return r
}
